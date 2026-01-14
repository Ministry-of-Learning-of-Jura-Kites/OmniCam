import functools
from state import CameraState, State
import numpy as np
from dataclasses import replace
from utils import center_of_face, look_at_quaternion
import quaternion


@functools.cache
def state_vector_dim():
    face = np.array(
        [
            [-1.0, -1.0, -1.0],
            [-1.0, -1.0, 1.0],
            [-1.0, 1.0, 1.0],
            [-1.0, 1.0, -1.0],
        ]
    )
    vec = state_to_vector(
        state=State(
            [
                CameraState(
                    face=face,
                    pos=np.array([0, 0, 0]),
                    angle=quaternion.from_float_array([0, 0, 0, 0]),
                    pixels=np.array([1920, 1080]),
                    vfov=70,
                )
            ],
            scale=1,
            gltf=None,
        )
    )

    return len(vec)


def state_to_vector(state: State):
    """Flattens State into a 1D numpy array."""
    vec = []
    for cam in state.cameras:
        vec.extend(cam.pos)  # Just 3 params: x, y, z
    return np.array(vec)


def vector_to_state(vec, template_state: State):
    """Reconstructs State from a 1D numpy array."""
    new_cameras = []
    idx = 0
    for i in range(len(template_state.cameras)):
        pos = vec[idx : idx + 3]

        # Calculate Look-At rotation automatically
        face_center = center_of_face(template_state.cameras[i].face)
        direction = face_center - pos
        # Use your existing utility to keep the camera pointed at the target
        angle = look_at_quaternion(direction)

        cam = replace(template_state.cameras[i], pos=pos, angle=angle)
        new_cameras.append(cam)
        idx += 3

    return State(
        cameras=new_cameras, scale=template_state.scale, gltf=template_state.gltf
    )
