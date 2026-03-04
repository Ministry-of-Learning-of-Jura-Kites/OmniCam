import copy
import math
import os
import time
from scipy.spatial.distance import cdist
from cost_functions import angle_cost, total_cost
from algorithms.differential_evolution import optimize_de
import numpy as np
from state import CameraConfiguration, CameraState, State, render_from_state
import quaternion
from utils import (
    center_of_face,
    get_seeded_color_rgb,
    look_at_quaternion,
    normal_vec_of_face,
)
from env import env_settings
import logging
import vtk

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

if env_settings.dev_mode:
    import pyvista as pv
    from pyvistaqt import BackgroundPlotter
    import pyvistaqt

    pl = BackgroundPlotter()


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


faces = []
faces.append(
    create_arbitrary_face(
        center=[22.0, -0.65, 0.0],
        width=3.2,
        height=2.3,
        normal=[-1, 0, 0],  # Facing 'inward' towards the room
    )
)

faces.append(
    create_arbitrary_face(
        center=[22.0, -0.65, 4.0], width=3.2, height=2.3, normal=[-1, 0, 0]
    )
)

faces.append(
    create_arbitrary_face(
        center=[9.0, 1.0, 4.0],
        width=2.82,
        height=2.0,
        normal=[1, 0, 1],  # 45-degree tilt
    )
)

faces.append(
    create_arbitrary_face(
        center=[15, 0, 0], width=4, height=3, normal=[1, 0, 1]  # Points diagonally
    )
)

faces.append(
    create_arbitrary_face(center=[11, 2, 2], width=5, height=5, normal=[0, 0, 1])
)

faces.append(
    create_arbitrary_face(center=[5, 5, 5], width=2, height=2, normal=[0.3, 0, -0.5])
)

gltf = (
    pv.read("~/Downloads/omnicam/cpn-lidar.glb")
    .combine()
    .extract_surface()
    .triangulate()
)

gltf_locator = vtk.vtkStaticCellLocator()
gltf_locator.SetDataSet(gltf)
gltf_locator.BuildLocator()

default_cam_config = CameraConfiguration(
    pixels=np.array([1920, 1080]),
    vfov=50,
)

extra_cam_config = CameraConfiguration(
    pixels=np.array([3840, 2160]),
    vfov=90,
)

default_cam = CameraState(
    pos=np.array([0, 0, 0]),
    angle=quaternion.from_rotation_vector([0, 0, 0]),
    faces=None,
    center_of_faces=None,
    camera_config=default_cam_config,
)
extra_cam = CameraState(
    pos=np.array([0, 0, 0]),
    angle=quaternion.from_rotation_vector([0, 0, 0]),
    faces=None,
    center_of_faces=None,
    camera_config=extra_cam_config,
)


state = State(
    faces=faces,
    face_centers=list(map(center_of_face, faces)),
    cameras=[
        copy.copy(default_cam),
        copy.copy(extra_cam),
        copy.copy(extra_cam),
    ],
    scale=1 / 2.5,
    gltf=gltf,
    gltf_locator=gltf_locator,
)

from scipy.spatial.distance import cdist


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
    face_to_cam = np.argmin(dist_mat, axis=1)
    assignments = [[] for _ in range(num_cameras)]

    for face_idx, cam_idx in enumerate(face_to_cam):
        assignments[cam_idx].append(state.faces[face_idx])

    # 4. Handle "Empty Camera" & Update State
    for i, cam in enumerate(state.cameras):
        if not assignments[i]:
            closest_face_idx = np.argmin(dist_mat[:, i])
            assignments[i].append(state.faces[closest_face_idx])

        cam.faces = assignments[i]
        if assignments[i]:
            cam.center_of_faces = np.mean(
                [center_of_face(f) for f in assignments[i]], axis=0
            )

    return state


def init_3d_scene(pl: pyvistaqt.BackgroundPlotter | None, state: State):
    if pl == None:
        return
    pl.add_mesh(state.gltf)
    pl.show_grid(color="gray", location="outer")
    face_mesh: pv.PolyData | None = None
    for i, face in enumerate(state.faces):
        # face_center = center_of_face(face)

        faces = np.hstack([[4, 0, 1, 2, 3]])
        face_mesh = pv.PolyData(face, faces=faces)
        color = get_seeded_color_rgb(i)
        pl.add_mesh(face_mesh, color=color)

    for camera in state.cameras:
        # camera.angle = look_at_quaternion(face_center - camera.pos)

        arrow = pv.Arrow(start=(0, 0, 0), direction=(1.0, 0.0, 0.0))
        camera.meshes.camera_actor = pl.add_mesh(arrow, color=color)
        silhouette_actor = pl.add_silhouette(
            arrow,
            color="white",
            line_width=8.0,
        )
        camera.meshes.camera_silhouette_actor = silhouette_actor

        temp_cam = pv.Camera()
        temp_cam.position = np.array([0, 0, 0])
        temp_cam.clipping_range = (0.1, 10.0)
        temp_cam.focal_point = temp_cam.position + np.array([1, 0, 0])
        temp_cam.up = (0, 1, 0)
        temp_cam.view_angle = camera.camera_config.vfov
        aspect = camera.camera_config.pixels[0] / camera.camera_config.pixels[1]
        frustum = temp_cam.view_frustum(aspect)
        camera.meshes.frustum_actor = pl.add_mesh(
            frustum, color=color, style="wireframe", opacity=0.5, line_width=2
        )

    pl.camera.up = (0, 1, 0)
    pl.camera_set = True
    pl.add_axes()
    pl.show_axes()
    pl.reset_camera(render=True, bounds=face_mesh.bounds)
    pl.enable_trackball_style()


def main():
    global state
    if env_settings.dev_mode:
        init_3d_scene(pl, state)
        render_from_state(pl, state)
        pl.show()

    seed = 2000

    state = assign_faces(state, seed)

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


if __name__ == "__main__":
    main()
