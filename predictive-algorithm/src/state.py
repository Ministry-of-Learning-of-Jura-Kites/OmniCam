from pyvistaqt import BackgroundPlotter
from utils import get_seeded_color_rgb
from dataclasses import dataclass
import time
from typing import List

import numpy as np
import pyvista as pv
from basic_types import Array2x2, Array4x3, Array3
import quaternion


@dataclass
class CameraState:
    face: Array4x3
    pos: Array3
    angle: quaternion.quaternion
    pixels: Array2x2
    vfov: float


@dataclass
class State:
    cameras: List[CameraState]
    scale: float  # virtual metre/metre


class InteractiveOptimizerPlotter:
    def __init__(self, plotter: BackgroundPlotter, initial_state: State):
        self.plotter = plotter

        self.plotter.show_axes()
        self.plotter.enable_trackball_style()

        self.state = initial_state
        self.skip_iterations = 0
        self.running = False

        # UI Toggles
        self.plotter.add_key_event("n", self.next_step)  # Press 'n' for next iteration
        self.plotter.add_key_event("c", self.continue_run)  # Press 'c' to run freely

        self.plotter.show()

    def next_step(self):
        self.skip_iterations = 0  # Stop at the very next iteration
        self.running = False

    def continue_run(self):
        self.running = True

    def update(self, state: State, iteration: int):
        breakpoint()
        render_from_state(self.plotter, state)
        # BLOCKING LOGIC: If not in 'running' mode, loop here
        # while not self.running and self.skip_iterations <= 0:
        #     self.plotter.update()
        #     time.sleep(0.01)  # Prevent CPU hogging


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
