import math
import os

import pyvistaqt
from cost_functions import angle_cost
from algorithms.particle_swarm_opt import optimize_pso
from algorithms.differential_evolution import optimize_de
import numpy as np
from state import CameraState, State, render_from_state
import pyvista as pv
from pyvistaqt import BackgroundPlotter
import quaternion
from utils import get_seeded_color_rgb

pl = BackgroundPlotter()

face = np.array(
    [
        [5.0, -1.0, -1.0],
        [5.0, -1.0, 1.0],
        [5.0, 1.0, 1.0],
        [5.0, 1.0, -1.0],
    ]
)
state = State(
    [
        CameraState(
            face=face,
            pos=np.array([19, 0, 0]),
            angle=quaternion.from_rotation_vector([0, -np.pi / 4, 0]),
            pixels=np.array([1920, 1080]),
            vfov=70,
        )
    ],
    scale=1,
    gltf=pv.read("~/Downloads/oxygenai models/cpn-lidar.glb")
    .combine()
    .extract_surface()
    .triangulate(),
)


def init_state(pl: pyvistaqt.BackgroundPlotter, state: State):
    # pl.add_mesh(state.gltf)
    for i, camera in enumerate(state.cameras):
        color = get_seeded_color_rgb(i)

        arrow = pv.Arrow(start=(0, 0, 0), direction=(0.0, 0.0, 1.0))
        camera.meshes.camera_actor = pl.add_mesh(arrow, color=color)
        silhouette_actor = pl.add_silhouette(
            arrow,
            color="white",
            line_width=8.0,
        )
        camera.meshes.camera_silhouette_actor = silhouette_actor

        cone_length = 10.0
        # Radius = length * tan(half_angle)
        cone_radius = cone_length * np.tan(np.radians(camera.vfov / 2))

        view_cone = pv.Cone(
            center=(0, 0, 1 + cone_length / 2),  # Offset center so apex is at origin
            direction=(0, 0, -1),  # Pointing along Z
            height=cone_length,
            radius=cone_radius,
            resolution=32,
        )

        # Add the cone with transparency (opacity)
        camera.meshes.cone_actor = pl.add_mesh(
            view_cone, color=color, opacity=0.2, style="surface"
        )

        faces = np.hstack([[4, 0, 1, 2, 3]])
        face_mesh = pv.PolyData(camera.face, faces=faces)
        pl.add_mesh(face_mesh, color=color)
        camera.meshes.face_mesh = face_mesh

    pl.add_axes()
    pl.show_axes()
    pl.enable_trackball_style()


def main():
    init_state(pl, state)
    render_from_state(pl, state)
    pl.show()

    breakpoint()

    # from cost_functions import total_cost
    # print(total_cost(state))
    # breakpoint()

    final_state = optimize_pso(
        state,
        pl,
        # None,
    )
    # final_state = optimize_de(
    #     state,
    #     gltf
    # )

    render_from_state(pl, final_state)
    breakpoint()


if __name__ == "__main__":
    main()
