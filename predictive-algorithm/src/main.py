import math
import os
from cost_functions import angle_cost
from algorithms.particle_swarm_opt import optimize_camera_pso
import numpy as np
from state import CameraState, State, render_from_state
import pyvista as pv
from pyvistaqt import BackgroundPlotter
import quaternion

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
)


def main():
    render_from_state(pl, state)
    pl.show()

    # from cost_functions import total_cost

    # print(total_cost(state))

    # breakpoint()

    optimize_camera_pso(
        state,
        pl,
    )

    # embed()


if __name__ == "__main__":
    main()
