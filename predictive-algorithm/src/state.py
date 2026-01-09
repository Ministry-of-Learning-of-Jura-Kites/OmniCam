from dataclasses import dataclass
from typing import List
from basic_types import Array2x2, Array4x3, Array3
import quaternion


@dataclass
class CameraState:
    face: Array4x3
    pos: Array3
    angle: quaternion.quaternion
    pixels: Array2x2
    fov: float


@dataclass
class State:
    cameras: List[CameraState]
    scale: float  # virtual metre/metre
