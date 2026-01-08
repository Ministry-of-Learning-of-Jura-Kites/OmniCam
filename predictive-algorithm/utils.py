import numpy as np
import quaternion
from basic_types import Array4x3, Array3


def center_of_face(face: Array4x3) -> Array3:
    return face.mean()


def angle_from_point(point: Array3, pos: Array3, angle: quaternion.quaternion):
    local_front_vector = np.array([0.0, 0.0, 1.0])

    # Rotate the local vector to get the world-space front vector
    world_front_vector = quaternion.rotate_vectors(angle, local_front_vector)

    to_point_vec = point - pos

    if np.linalg.norm(world_front_vector) * np.linalg.norm(to_point_vec) == 0:
        return (
            (world_front_vector * to_point_vec)
            / np.linalg.norm(world_front_vector)
            * np.linalg.norm(to_point_vec)
        )

    return 0
