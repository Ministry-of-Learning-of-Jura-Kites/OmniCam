from pyvistaqt import BackgroundPlotter
from utils import get_seeded_color_rgb
from dataclasses import dataclass, field
import time
from typing import Any, List, Optional
import vtk
import numpy as np
import pyvista as pv
from basic_types import Array2x2, Array4x3, Array3
import quaternion
from scipy.spatial.transform import Rotation as R


@dataclass
class CameraMesh:
    face_mesh: Optional[pv.PolyData] = None
    camera_actor: Optional[Any] = None
    camera_silhouette_actor: Optional[Any] = None
    frustum_actor: Optional[Any] = None


@dataclass
class CameraConfiguration:
    pixels: Array2x2
    vfov: float


@dataclass
class CameraState:
    # face: Array4x3
    pos: Array3
    angle: quaternion.quaternion
    # pixels: Array2x2
    # vfov: float
    meshes: CameraMesh = field(default_factory=CameraMesh)
    camera_config: CameraConfiguration = field(default_factory=CameraConfiguration)

    def forward_vector(self) -> Array3:
        r = R.from_quat(quaternion.as_float_array(self.angle))
        return r.apply(np.array([1, 0, 0]))


@dataclass
class State:
    faces: List[Array4x3]
    face_centers: List[Array3]
    cameras: List[CameraState]
    gltf: pv.DataObject
    gltf_locator: vtk.vtkStaticCellLocator
    scale: float  # virtual metre/metre


class InteractiveOptimizerPlotter:
    def __init__(self, plotter: BackgroundPlotter, initial_state: State):
        self.plotter = plotter

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
        render_from_state(self.plotter, state)
        # breakpoint()

    # def update(self, state: State, iteration: int):
    #     breakpoint()
    #     render_from_state(self.plotter, state)


def render_from_state(pl: pv.Plotter, state: State):
    # pl.clear()
    for i, camera in enumerate(state.cameras):
        rot_mat = quaternion.as_rotation_matrix(camera.angle)

        transform = np.eye(4)
        transform[:3, :3] = rot_mat
        transform[:3, 3] = camera.pos
        vtk_matrix = vtk.vtkMatrix4x4()
        for row in range(4):
            for col in range(4):
                vtk_matrix.SetElement(row, col, transform[row, col])
        camera.meshes.camera_actor.SetUserMatrix(vtk_matrix)
        camera.meshes.camera_silhouette_actor.SetUserMatrix(vtk_matrix)
        camera.meshes.frustum_actor.SetUserMatrix(vtk_matrix)
