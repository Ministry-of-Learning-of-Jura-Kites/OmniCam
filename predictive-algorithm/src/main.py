import math
import os
import time

import pyvistaqt
from cost_functions import angle_cost, total_cost
from algorithms.particle_swarm_opt import optimize_pso
from algorithms.differential_evolution import optimize_de
import numpy as np
from state import CameraState, State, render_from_state
import pyvista as pv
from pyvistaqt import BackgroundPlotter
import quaternion
from utils import center_of_face, get_seeded_color_rgb, look_at_quaternion
import logging
import vtk

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

pl = BackgroundPlotter()

face = np.array(
    [
        [15.0, -1.8, -1.6],
        [15.0, -1.8, 1.6],
        [15.0, 0.5, 1.6],
        # [24.0, 3.0, 3.0],
        [15.0, 0.5, -1.6],
    ]
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

state = State(
    [
        CameraState(
            face=face,
            pos=np.array([0, 5, 20]),
            angle=quaternion.from_rotation_vector([0, 0, 0]),
            pixels=np.array([1920, 1080]),
            vfov=70,
        )
    ],
    scale=1 / 2.5,
    gltf=gltf,
    gltf_locator=gltf_locator,
)


def init_state(pl: pyvistaqt.BackgroundPlotter, state: State):
    pl.add_mesh(state.gltf)
    pl.show_grid(color="gray", location="outer")
    for i, camera in enumerate(state.cameras):
        color = get_seeded_color_rgb(i)

        face_center = center_of_face(camera.face)

        camera.angle = look_at_quaternion(face_center - camera.pos)

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
        temp_cam.view_angle = camera.vfov
        aspect = camera.pixels[0] / camera.pixels[1]
        frustum = temp_cam.view_frustum(aspect)
        camera.meshes.frustum_actor = pl.add_mesh(
            frustum, color=color, style="wireframe", opacity=0.5, line_width=2
        )

        faces = np.hstack([[4, 0, 1, 2, 3]])
        face_mesh = pv.PolyData(camera.face, faces=faces)
        pl.add_mesh(face_mesh, color=color)
        camera.meshes.face_mesh = face_mesh

    pl.camera.up = (0, 1, 0)
    pl.camera_set = True
    pl.add_axes()
    pl.show_axes()
    pl.reset_camera(render=True, bounds=state.cameras[0].meshes.face_mesh.bounds)
    pl.enable_trackball_style()


def main():
    init_state(pl, state)
    render_from_state(pl, state)
    pl.show()

    breakpoint()

    start_time = time.perf_counter()
    # from cost_functions import total_cost
    # print(total_cost(state))
    # breakpoint()

    # final_state = optimize_pso(
    #     state,
    #     # pl,
    #     None,
    # )
    final_state = optimize_de(state)

    end_time = time.perf_counter()
    elapsed_time = end_time - start_time
    print(f"Elapsed time: {elapsed_time:.4f} seconds")

    render_from_state(pl, final_state)

    total, result = total_cost(final_state)
    print("cost distribution: ", result)
    print("total cost: ", total)
    breakpoint()


if __name__ == "__main__":
    main()
