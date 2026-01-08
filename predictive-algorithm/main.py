import os
from IPython import embed
import numpy as np
from state import CameraState, State
import pyvista as pv
from pyvistaqt import BackgroundPlotter
import quaternion


def render_from_state(pl: pv.Plotter, state: State):
    pl.clear()
    for camera in state.cameras:
        local_front_vector = np.array([0.0, 0.0, 1.0])

        # Rotate the local vector to get the world-space front vector
        world_front_vector = quaternion.rotate_vectors(camera.angle, local_front_vector)
        pl.add_mesh(pv.Arrow(start=camera.pos, direction=world_front_vector))


def main():
    face = np.array([[-1, -1, -1], [-1, 1, 1], [-1, -1, 1], [-1, 1, -1]])
    state = State(
        [
            CameraState(
                face, np.array([0, 0, 0]), quaternion.from_euler_angles([0, 0, 0])
            )
        ]
    )

    pl = BackgroundPlotter()
    render_from_state(pl, state)
    pl.show_axes()
    pl.enable_trackball_actor_style()
    pl.show()

    embed()


if __name__ == "__main__":
    main()
