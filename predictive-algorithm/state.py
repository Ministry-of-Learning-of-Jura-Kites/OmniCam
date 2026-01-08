from dataclasses import dataclass
from typing import List
from basic_types import Array4x3, Array3
import quaternion


@dataclass
class CameraState:
    face: Array4x3
    pos: Array3
    angle: quaternion.quaternion


@dataclass
class State:
    cameras: List[CameraState]
