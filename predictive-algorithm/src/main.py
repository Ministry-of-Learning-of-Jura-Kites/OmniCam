import asyncio
import math
from os import path
import signal
import sys
import time
import traceback
from typing import List, Tuple
import uuid
import nats
from pydantic import BaseModel, ValidationError
from scipy.spatial.distance import cdist
from cost_functions import total_cost
from google.protobuf.json_format import MessageToJson
import messages.protobufs.camera_pb2 as cam_pb
import messages.protobufs.optimization_pb2 as opt_pb
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
from nats.aio.msg import Msg

# ===================== TELEMETRY ADDED =====================
from pkg.telemetry import init_telemetry
from opentelemetry import trace, propagate
from opentelemetry.trace import SpanKind, StatusCode

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("omnicam-algo")


# ===================== MODELS =====================
class ReqCameraConfiguration(BaseModel):
    pixels: Tuple[float, float]
    vfov: float
    name: str
    amount: int


class OptimizeRequest(BaseModel):
    faces: List[List[Tuple[float, float, float]]]
    cam_configs: List[ReqCameraConfiguration]
    scale: float
    job_id: str
    model_id: str
    project_id: str


# ===================== FACE TRANSFORM (UNCHANGED) =====================
def transform_faces(faces: List[List[Tuple[float, float, float]]]) -> Array4x3:
    cleaned = []
    for face in faces:
        if face is None or len(face) < 3:
            continue

        arr = np.array(face, dtype=np.float64)

        if np.isnan(arr).any() or np.allclose(arr[0], arr[-1]):
            continue

        cleaned.append(arr)

    return cleaned


# ===================== CAMERA TRANSFORM (UNCHANGED) =====================
def transform_cameras(raw_cam_configs: List[ReqCameraConfiguration]):
    cameras = []
    for raw in raw_cam_configs:
        cam_config = CameraConfiguration(
            pixels=raw.pixels,
            vfov=raw.vfov,
            name=raw.name,
        )

        for _ in range(raw.amount):
            cameras.append(
                CameraState(
                    faces=None,
                    pos=[5, 0, 0],
                    angle=quaternion.from_vector_part([0, 0, 0, 1]),
                    center_of_faces=None,
                    camera_config=cam_config,
                    name=raw.name,
                )
            )
    return cameras


# ===================== ASSIGNMENT LOGIC (UNCHANGED) =====================
def assign_faces(state: State, seed: int):
    num_faces = len(state.faces)
    num_cameras = len(state.cameras)

    if num_faces == 0 or num_cameras == 0:
        return state

    face_centers = np.array([center_of_face(f) for f in state.faces])
    face_normals = np.array([normal_vec_of_face(f) for f in state.faces])

    rng = np.random.default_rng(seed)

    seeds_idx = [rng.integers(0, num_faces)]
    for _ in range(1, num_cameras):
        dist_sq = np.min(cdist(face_centers, face_centers[seeds_idx]), axis=1) ** 2
        probs = dist_sq / dist_sq.sum()
        seeds_idx.append(rng.choice(num_faces, p=probs))

    seed_centers = face_centers[seeds_idx]
    seed_normals = face_normals[seeds_idx]

    dist_mat = cdist(face_centers, seed_centers, metric="euclidean")

    for c_idx, cam in enumerate(state.cameras):
        vfov_rad = math.radians(cam.camera_config.vfov)
        v_res = cam.camera_config.pixels[1]

        vecs = face_centers - seed_centers[c_idx]
        distances = np.linalg.norm(vecs, axis=1) + 1e-6

        cos_sim = np.dot(face_normals, seed_normals[c_idx])
        norm_penalty = np.where(cos_sim < 0, 100.0, 1.0 - cos_sim)

        face_scale = state.scale
        angular_size = 2 * np.arctan(face_scale / (2 * distances))

        fov_penalty = np.where(angular_size > vfov_rad, 10.0, 1.0)

        pixels_covered = (angular_size / vfov_rad) * v_res
        res_penalty = np.where(pixels_covered < 50, 5.0, 1.0)

        dist_mat[:, c_idx] *= (1.0 + norm_penalty * 5.0) * fov_penalty * res_penalty

    face_to_cam = state.face_to_cam
    assignments = [[] for _ in range(num_cameras)]

    face_to_cam_dist = np.argmin(dist_mat, axis=1)

    for i, cam_idx in enumerate(face_to_cam_dist):
        assignments[cam_idx].append(state.faces[i])
        face_to_cam[int(i)] = int(cam_idx)

    for i, cam in enumerate(state.cameras):
        if not assignments[i]:
            closest = np.argmin(dist_mat[:, i])
            assignments[i].append(state.faces[closest])
            face_to_cam[int(closest)] = int(i)

        cam.faces = assignments[i]
        cam.center_of_faces = np.mean(
            [center_of_face(f) for f in assignments[i]], axis=0
        )
        cam.angle = look_at_quaternion(cam.center_of_faces - cam.pos)

    state.face_to_cam = face_to_cam
    return state


# ===================== OPTIMIZE (UNCHANGED LOGIC) =====================
def optimize(req: OptimizeRequest, seed: int = 2000) -> State:
    pl = None
    if env_settings.dev_mode:
        from pyvistaqt import BackgroundPlotter

        pl = BackgroundPlotter()

    raw = pv.read(
        path.join(
            env_settings.model_file_path,
            "3d_models",
            req.project_id,
            req.model_id + ".glb",
        )
    )

    combined = raw.combine()
    surface = combined.extract_surface()

    gltf = surface.clean(tolerance=1e-5).triangulate()

    if not gltf.is_all_triangles:
        gltf = gltf.extract_cells_by_type(vtk.VTK_TRIANGLE)

    locator = vtk.vtkStaticCellLocator()
    locator.SetDataSet(gltf)
    locator.BuildLocator()

    faces = transform_faces(req.faces)
    cameras = transform_cameras(req.cam_configs)

    state = State(
        faces=faces,
        face_to_cam={},
        face_centers=list(map(center_of_face, faces)),
        cameras=cameras,
        scale=req.scale,
        gltf=gltf,
        gltf_locator=locator,
    )

    if len(cameras) > len(faces):
        return state

    state = assign_faces(state, seed)

    if env_settings.dev_mode:
        init_3d_scene(pl, state)
        render_from_state(pl, state)
        pl.show()

    start = time.perf_counter()
    final_state, _ = optimize_de(state, seed)
    elapsed = time.perf_counter() - start

    logger.info(f"optimization_time={elapsed:.4f}")
    logger.info(f"total_cost={total_cost(final_state, True)}")

    if env_settings.dev_mode:
        render_from_state(pl, final_state)
        pl.close()

    return final_state


# ===================== PROTO =====================
def cam_state_to_proto(cam_state: CameraState) -> cam_pb.Camera:
    cam = cam_pb.Camera()
    cam.name = cam_state.name
    cam.id = str(uuid.uuid4())

    cam.pos_x, cam.pos_y, cam.pos_z = cam_state.pos

    cam.angle_w = cam_state.angle.w
    cam.angle_x = cam_state.angle.x
    cam.angle_y = cam_state.angle.y
    cam.angle_z = cam_state.angle.z

    config = cam_state.camera_config
    cam.fov = config.vfov
    cam.width_res = config.pixels[0]
    cam.height_res = config.pixels[1]

    cam.frustum_length = 10
    return cam


# ===================== MAIN (TRACE ADDED ONLY) =====================
async def main():
    tracer = init_telemetry()

    async def message_handler(msg: Msg):
        carrier = dict(msg.headers) if msg.headers else {}
        ctx = propagate.extract(carrier)

        with tracer.start_as_current_span(
            "optimization.process",
            context=ctx,
            kind=SpanKind.CONSUMER,
        ) as span:

            logger.info("job_received", extra={"subject": msg.subject})

            try:
                payload = OptimizeRequest.model_validate_json(msg.data)
                span.set_attribute("job_id", payload.job_id)

                logger.info("optimization_started", extra={"job_id": payload.job_id})

                result = optimize(payload)

                logger.info("optimization_finished", extra={"job_id": payload.job_id})

                response = opt_pb.OptimizationEventResp(
                    success_resp=opt_pb.SuccessOptimizationEventResp(
                        cameras=[cam_state_to_proto(c) for c in result.cameras],
                    ),
                    job_id=payload.job_id,
                )

                span.set_status(StatusCode.OK)

            except ValidationError as e:
                span.set_status(StatusCode.ERROR, str(e))
                logger.error("validation_error", extra={"error": str(e)})

                response = opt_pb.OptimizationEventResp(
                    error_resp=opt_pb.EventError(internal_error=opt_pb.SimpleError()),
                    job_id="",
                )

            except Exception as e:
                span.set_status(StatusCode.ERROR, str(e))
                span.record_exception(e)

                logger.exception("optimization_failed")

                response = opt_pb.OptimizationEventResp(
                    error_resp=opt_pb.EventError(internal_error=opt_pb.SimpleError()),
                    job_id="",
                )

            finally:
                await msg.respond(MessageToJson(response, indent=0).encode("utf-8"))

    nc = await nats.connect(env_settings.nats_url)

    await nc.subscribe(
        subject=env_settings.req_topic_pattern.format(jobId="*"),
        queue=env_settings.req_topic_queue,
        cb=message_handler,
    )

    await asyncio.Future()


def handle_sigterm(signum, frame):
    sys.exit(0)


signal.signal(signal.SIGTERM, handle_sigterm)

if __name__ == "__main__":
    asyncio.run(main())
