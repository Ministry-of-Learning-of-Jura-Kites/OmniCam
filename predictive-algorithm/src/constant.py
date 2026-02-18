from enum import Enum
from typing import Literal


class VectorSerialization(Enum):
    CARTESIAN = "CARTESIAN"
    SPHERICAL = "SPHERICAL"


BIG_M = 1e10
VECTOR_SERIALIZATION: VectorSerialization = VectorSerialization.CARTESIAN
