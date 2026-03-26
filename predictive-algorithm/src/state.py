import math
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional
import vtk
import numpy as np
from basic_types import Array2, Array4x3, Array3
import quaternion
import pyvista as pv


@dataclass
class CameraMesh:
    face_mesh: Optional["pv.PolyData"] = None
    camera_actor: Optional[Any] = None
    camera_silhouette_actor: Optional[Any] = None
    frustum_actor: Optional[Any] = None


@dataclass
class CameraConfiguration:
    pixels: Array2
    vfov: float
    name: str

    def get_hfov(self):
        """
        Calculates Horizontal FOV using the trigonometric relationship:
        HFOV = 2 * arctan(tan(VFOV/2) * aspect_ratio)
        """
        width = self.pixels[0]
        height = self.pixels[1]
        aspect_ratio = width / height

        # 1. Convert VFOV to radians and find the half-angle
        vfov_rad = math.radians(self.vfov)

        # 2. Calculate the tangent of the half-angle
        tan_half_vfov = math.tan(vfov_rad / 2)

        # 3. Scale by aspect ratio and find the inverse tangent
        hfov_rad = 2 * math.atan(tan_half_vfov * aspect_ratio)

        # 4. Convert back to degrees
        return math.degrees(hfov_rad)


@dataclass
class CameraState:
    # face: Array4x3
    faces: List[Array4x3] | None
    pos: Array3
    angle: quaternion.quaternion
    name: str
    # pixels: Array2x2
    # vfov: float
    center_of_faces: Array3 | None
    meshes: CameraMesh = field(default_factory=CameraMesh)
    camera_config: CameraConfiguration = field(default_factory=CameraConfiguration)

    def forward_vector(self) -> Array3:
        return quaternion.rotate_vectors(self.angle, np.array([1, 0, 0]))


@dataclass
class State:
    faces: List[Array4x3]
    face_to_cam: Dict[int, int]
    face_centers: List[Array3]
    cameras: List[CameraState]
    gltf: pv.PolyData
    gltf_locator: vtk.vtkStaticCellLocator
    scale: float  # real-life metre / virtual metre
