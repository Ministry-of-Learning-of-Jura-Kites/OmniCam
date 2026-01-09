import math
import os
from cost_functions import angle_cost
import numpy as np
from state import CameraState, State
import pyvista as pv
from pyvistaqt import BackgroundPlotter
import quaternion
from utils import get_seeded_color_rgb


def render_from_state(pl: pv.Plotter, state: State):
    pl.clear()
    for i, camera in enumerate(state.cameras):
        color = get_seeded_color_rgb(i)

        local_front_vector = np.array([0.0, 0.0, 1.0])

        # Rotate the local vector to get the world-space front vector
        world_front_vector = quaternion.rotate_vectors(camera.angle, local_front_vector)
        arrow = pv.Arrow(start=camera.pos, direction=world_front_vector)
        silhouette = dict(
            color="white",
            line_width=8.0,
        )
        pl.add_mesh(arrow, color=color, silhouette=silhouette)

        faces = np.hstack([[4, 0, 1, 2, 3]])
        pl.add_mesh(pv.PolyData(camera.face, faces=faces), color=color)


pl = BackgroundPlotter()
face = np.array(
    [
        [-1.0, -1.0, -1.0],
        [-1.0, -1.0, 1.0],
        [-1.0, 1.0, 1.0],
        [-1.0, 1.0, -1.0],
    ]
)
state = State(
    [
        CameraState(
            face,
            np.array([0, 0, 0]),
            quaternion.from_float_array(
                [
                    0.022,
                    0.708,
                    0.443,
                    -0.550,
                ]
            ),
        )
    ]
)


def main():

    render_from_state(pl, state)
    pl.show_axes()
    pl.enable_trackball_style()

    pl.show()
    # embed()


if __name__ == "__main__":
    main()
