from typing import Annotated, List, Literal, Tuple, TypeVar

from numpy import typing as npt
import numpy as np

DType = TypeVar("DType", bound=np.generic)

Array4 = Annotated[npt.NDArray[DType], Literal[4]]
Array3 = Annotated[npt.NDArray[DType], Literal[3]]

Array4x3 = Annotated[npt.NDArray[DType], Literal[4, 3]]
