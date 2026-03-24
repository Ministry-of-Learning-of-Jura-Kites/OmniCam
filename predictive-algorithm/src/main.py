import asyncio
import json
import math
import time
from typing import List, Tuple
import uuid
from pydantic import BaseModel, ValidationError
import redis.asyncio as redis
from scipy.spatial.distance import cdist
from cost_functions import total_cost
from algorithms.differential_evolution import optimize_de
import numpy as np
from state import CameraConfiguration, CameraState, State
import quaternion
from utils import (
    center_of_face,
    look_at_quaternion,
    normal_vec_of_face,
)
from env import env_settings
import logging
import vtk
from dev.visualization import init_3d_scene, render_from_state
import pyvista as pv
from basic_types import Array4x3

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


# TODO: Use protobufs(?)
class ReqCameraConfiguration(BaseModel):
    pixels: Tuple[float, float]
    vfov: float
    name: str
    amount: int


class OptimizeRequest(BaseModel):
    faces: List[List[Tuple[float, float, float]]]
    cam_configs: List[ReqCameraConfiguration]
    scale: float
    job_id: uuid.UUID
    model_id: uuid.UUID


class CameraResponse(BaseModel):
    id: str
    name: str
    angle_x: float
    angle_y: float
    angle_z: float
    angle_w: float
    pos_x: float
    pos_y: float
    pos_z: float
    fov: float
    width_res: int
    height_res: int


class OptimizeResponse(BaseModel):
    cameras: List[CameraResponse]


def create_arbitrary_face(center, width, height, normal):
    """
    Generates a rectangular face centered at 'center' with a specific 'normal'.
    'normal' defines the tilt/rotation.
    """
    center = np.array(center)
    normal = np.array(normal) / np.linalg.norm(normal)

    # Create an orthogonal coordinate system around the normal
    # 1. Pick a temporary vector that isn't parallel to the normal
    up = np.array([0, 1, 0]) if abs(normal[1]) < 0.9 else np.array([1, 0, 0])

    # 2. Calculate local X (right) and local Y (up) axes for the face
    right = np.cross(up, normal)
    right /= np.linalg.norm(right)
    local_up = np.cross(normal, right)

    hw = width / 2.0
    hh = height / 2.0

    # 3. Define the 4 corners relative to center using local axes
    p1 = center - (right * hw) - (local_up * hh)
    p2 = center + (right * hw) - (local_up * hh)
    p3 = center + (right * hw) + (local_up * hh)
    p4 = center - (right * hw) + (local_up * hh)

    return np.array([p1, p2, p3, p4])


def assign_faces(state: State, seed: int):
    num_faces = len(state.faces)
    num_cameras = len(state.cameras)
    if num_cameras == 0 or num_faces == 0:
        return

    face_centers = np.array([center_of_face(f) for f in state.faces])
    face_normals = np.array([normal_vec_of_face(f) for f in state.faces])

    rng = np.random.default_rng(seed)

    # 1. K-Means++ Seed Initialization
    seeds_idx = [rng.integers(0, num_faces)]
    for _ in range(1, num_cameras):
        dist_sq = np.min(cdist(face_centers, face_centers[seeds_idx]), axis=1) ** 2
        probs = dist_sq / dist_sq.sum()
        seeds_idx.append(np.random.choice(num_faces, p=probs))

    seed_centers = face_centers[seeds_idx]
    seed_normals = face_normals[seeds_idx]

    # 2. Compute Hybrid Cost
    dist_mat = cdist(face_centers, seed_centers, metric="euclidean")

    for c_idx, cam in enumerate(state.cameras):
        # --- Physical Camera Constraints ---
        # Get camera-specific config
        vfov_rad = math.radians(cam.camera_config.vfov)
        v_res = cam.camera_config.pixels[1]

        # Vector from seed to all faces
        vecs = face_centers - seed_centers[c_idx]
        distances = np.linalg.norm(vecs, axis=1) + 1e-6

        # --- Normal Penalty ---
        cos_sim = np.dot(face_normals, seed_normals[c_idx])
        norm_penalty = np.where(cos_sim < 0, 100.0, 1.0 - cos_sim)

        # --- Resolution/FOV Penalty ---
        # Estimate angular size of the face from the camera seed
        # Assuming faces are somewhat uniform, we use a characteristic scale
        face_scale = state.scale
        angular_size = 2 * np.arctan(face_scale / (2 * distances))

        # FOV Fit: If angular size > vfov, the face won't fit in one frame
        fov_penalty = np.where(angular_size > vfov_rad, 10.0, 1.0)

        # Resolution Quality: How many pixels does the face cover?
        # We want faces to cover a reasonable % of resolution.
        # Too few pixels = high penalty.
        pixels_covered = (angular_size / vfov_rad) * v_res
        res_penalty = np.where(pixels_covered < 50, 5.0, 1.0)  # Penalty if < 50px

        # Combine costs
        dist_mat[:, c_idx] *= (1.0 + norm_penalty * 5.0) * fov_penalty * res_penalty

    # 3. Assign
    face_to_cam_dist = np.argmin(dist_mat, axis=1)
    face_to_cam_map = state.face_to_cam
    assignments = [[] for _ in range(num_cameras)]

    for face_idx, cam_idx in enumerate(face_to_cam_dist):
        assignments[cam_idx].append(state.faces[face_idx])
        face_to_cam_map[int(face_idx)] = int(cam_idx)

    # 4. Handle "Empty Camera" & Update State
    for i, cam in enumerate(state.cameras):
        if not assignments[i]:
            closest_face_idx = np.argmin(dist_mat[:, i])
            assignments[i].append(state.faces[closest_face_idx])
            face_to_cam_map[int(closest_face_idx)] = int(i)

        cam.faces = assignments[i]
        if assignments[i]:
            cam.center_of_faces = np.mean(
                [center_of_face(f) for f in assignments[i]], axis=0
            )
            cam.angle = look_at_quaternion(cam.center_of_faces - cam.pos)

    state.face_to_cam = face_to_cam_map

    return state


# faces = []
# faces.append(
#     create_arbitrary_face(
#         center=[22.0, -0.65, 0.0],
#         width=3.2,
#         height=2.3,
#         normal=[-1, 0, 0],  # Facing 'inward' towards the room
#     )
# )

# faces.append(
#     create_arbitrary_face(
#         center=[22.0, -0.65, 4.0], width=3.2, height=2.3, normal=[-1, 0, 0]
#     )
# )

# faces.append(
#     create_arbitrary_face(
#         center=[9.0, 1.0, 4.0],
#         width=2.82,
#         height=2.0,
#         normal=[1, 0, 1],  # 45-degree tilt
#     )
# )

# faces.append(
#     create_arbitrary_face(
#         center=[15, 0, 0], width=4, height=3, normal=[1, 0, 1]  # Points diagonally
#     )
# )

# faces.append(
#     create_arbitrary_face(center=[11, 2, 2], width=5, height=5, normal=[0, 0, 1])
# )

# faces.append(
#     create_arbitrary_face(center=[5, 5, 5], width=2, height=2, normal=[0.3, 0, -0.5])
# )

# default_cam_config = CameraConfiguration(
#     pixels=np.array([1920, 1080]),
#     vfov=50,
# )

# extra_cam_config = CameraConfiguration(
#     pixels=np.array([3840, 2160]),
#     vfov=90,
# )

# default_cam = CameraState(
#     pos=np.array([0, 0, 0]),
#     angle=quaternion.from_rotation_vector([0, 0, 0]),
#     faces=None,
#     center_of_faces=None,
#     camera_config=default_cam_config,
# )
# extra_cam = CameraState(
#     pos=np.array([0, 0, 0]),
#     angle=quaternion.from_rotation_vector([0, 0, 0]),
#     faces=None,
#     center_of_faces=None,
#     camera_config=extra_cam_config,
# )


def transform_faces(faces: List[List[Tuple[float, float, float]]]) -> Array4x3:
    return [np.array(face, dtype=np.float64) for face in faces]


def transform_cameras(raw_cam_configs: List[ReqCameraConfiguration]):
    cameras = []
    for raw_cam_config in raw_cam_configs:
        cam_config = CameraConfiguration(
            pixels=raw_cam_config.pixels,
            vfov=raw_cam_config.vfov,
            name=raw_cam_config.name,
        )
        for _ in range(raw_cam_config.amount):
            cameras.append(
                CameraState(
                    faces=None,
                    pos=[5, 0, 0],
                    angle=quaternion.from_vector_part([0, 0, 0, 1]),
                    center_of_faces=None,
                    camera_config=cam_config,
                    name=raw_cam_config.name,
                )
            )

    return cameras


def optimize(req: OptimizeRequest) -> State:
    pl = None
    if env_settings.dev_mode:
        from pyvistaqt import BackgroundPlotter

        pl = BackgroundPlotter()

    gltf = (
        pv.read("~/Downloads/omnicam/cpn-lidar.glb")
        .combine()
        .extract_surface()
        .triangulate()
        .clean()
    )

    gltf_locator = vtk.vtkStaticCellLocator()
    gltf_locator.SetDataSet(gltf)
    gltf_locator.BuildLocator()

    faces = transform_faces(req.faces)
    cameras = transform_cameras(req.cam_configs)
    state = State(
        faces=faces,
        face_to_cam=dict(),
        face_centers=list(map(center_of_face, faces)),
        cameras=cameras,
        scale=req.scale,
        gltf=gltf,
        gltf_locator=gltf_locator,
    )

    seed = 2000

    num_faces = len(state.faces)
    num_cameras = len(state.cameras)

    if num_cameras > num_faces:
        return

    state = assign_faces(state, seed)

    if env_settings.dev_mode:
        init_3d_scene(pl, state)
        render_from_state(pl, state)
        pl.show()

    start_time = time.perf_counter()
    # from cost_functions import total_cost
    # print(total_cost(state))
    # breakpoint()

    # final_state = optimize_pso(
    #     state,
    #     # pl,
    #     None,
    # )
    final_state = optimize_de(state, seed)

    end_time = time.perf_counter()
    elapsed_time = end_time - start_time
    print(f"Elapsed time: {elapsed_time:.4f} seconds")

    if env_settings.dev_mode:
        render_from_state(pl, final_state)

    total = total_cost(final_state, True)
    print("total cost: ", total)

    if env_settings.dev_mode:
        pl.close()

    return final_state


def serialize_response(state: State):
    cameras = []
    for cam_state in state.cameras:
        angle_x = cam_state.angle.x
        angle_y = cam_state.angle.y
        angle_z = cam_state.angle.z
        angle_w = cam_state.angle.w
        pos_x, pos_y, pos_z = cam_state.pos
        cam = CameraResponse(
            id=str(uuid.uuid4()),
            name=cam_state.camera_config.name,
            angle_x=angle_x,
            angle_y=angle_y,
            angle_z=angle_z,
            angle_w=angle_w,
            pos_x=pos_x,
            pos_y=pos_y,
            pos_z=pos_z,
            fov=cam_state.camera_config.vfov,
            width_res=cam_state.camera_config.pixels[0],
            height_res=cam_state.camera_config.pixels[1],
        )

        cameras.append(cam)

    return cameras


r = redis.Redis(
    host=env_settings.redis_host, port=env_settings.redis_port, decode_responses=True
)


async def worker():
    print("Starting worker...")

    while True:
        # Read from the task stream (Blocking read)
        # 0 means wait indefinitely for a new message
        messages = await r.xread({env_settings.redis_req_topic: "0"}, count=1, block=0)

        for _stream, msgs in messages:
            for msg_id, data in msgs:
                try:
                    payload = OptimizeRequest.model_validate_json(data)

                    result_state = optimize(payload)

                    resp = serialize_response(result_state)

                    # Publish back to a result topic/stream
                    await r.xadd(
                        env_settings.redis_res_topic,
                        {
                            "job_id": payload.job_id,
                            "status": "error",
                            "data": json.dumps(resp),
                        },
                    )
                except ValidationError as e:
                    await r.xadd(
                        env_settings.redis_res_topic,
                        {
                            "job_id": payload.job_id,
                            "status": "error",
                            "error": f"Bad request {e}",
                        },
                    )
                except Exception:
                    await r.xadd(
                        env_settings.redis_res_topic,
                        {
                            "job_id": payload.job_id,
                            "status": "error",
                            "error": "Internal error",
                        },
                    )
                finally:
                    await r.xdel(env_settings.redis_req_topic, msg_id)


if __name__ == "__main__":
    asyncio.run(worker())

# Running -> OmniCam/predictive-algorithm/src$ uvicorn main:app --reload --port 8081
