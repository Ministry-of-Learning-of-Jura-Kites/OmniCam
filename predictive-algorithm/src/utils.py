import math
import random
from typing import Tuple
import numpy as np
import quaternion
from basic_types import Array4x3, Array3
from astropy.units import Quantity
import astropy.units as u
import pyvista as pv


def get_seeded_color_rgb(seed_value):
    """
    Generates a color (RGB tuple) from a specific seed.
    """
    # Seed the random number generator.
    # Passing the same seed value will ensure the same color is generated every time.
    random.seed(seed_value)

    # Generate random values for red, green, and blue components
    r = random.randint(0, 255)
    g = random.randint(0, 255)
    b = random.randint(0, 255)

    # Return the color as a tuple
    return (r, g, b)


def center_of_face(face: Array4x3) -> Array3:
    return np.mean(face, axis=0)


def angle_from_face_normal(
    face: Array4x3,
    pos: Array3,
    angle: quaternion.quaternion,
) -> Tuple[Quantity[u.radian], Quantity[u.radian]]:
    """Returns vertical and horizontal angles in radian unit"""

    face_center = np.mean(face, axis=0)
    angle_vec = quaternion.rotate_vectors(angle, [0, 0, 1])
    look_vec = face_center - pos

    horizontal_offset = math.atan2(look_vec[1], look_vec[0]) - math.atan2(
        angle_vec[1], angle_vec[0]
    )

    highest_face_center = np.array([face_center[0], face_center[1], face[:, 2].max()])
    look_vec = highest_face_center - pos
    dist_xy_look = math.sqrt(look_vec[0] ** 2 + look_vec[1] ** 2)
    dist_xy_cam = math.sqrt(angle_vec[0] ** 2 + angle_vec[1] ** 2)

    vert_look_angle = math.atan2(look_vec[2], dist_xy_look)
    vert_cam_angle = math.atan2(angle_vec[2], dist_xy_cam)

    vertical_offset = vert_look_angle - vert_cam_angle

    return (
        horizontal_offset * u.radian,
        vertical_offset * u.radian,
    )


# def angle_from_point(
#     point: Array3, pos: Array3, angle: quaternion.quaternion
# ) -> Tuple[Quantity[u.radian], Quantity[u.radian]]:
#     """Returns vertical and horizontal angles in radian unit"""
#     to_point_world = point - pos

#     if np.linalg.norm(to_point_world) == 0:
#         return 0.0, 0.0

#     # Rotate the world vector into LOCAL space
#     to_point_local = quaternion.rotate_vectors(angle.conjugate(), to_point_world)

#     # Assuming:
#     # Z-axis = Forward
#     # X-axis = Right
#     # Y-axis = Up
#     x, y, z = to_point_local

#     # Horizontal: Use Y and X (The floor plane)
#     # Since -X is Forward, we use -loc_x as the adjacent side
#     # and -loc_y (Right) or +loc_y (Left) as the opposite side
#     horizontal_angle = math.atan2(y, -x)

#     # Vertical: Angle above/below the XY plane
#     dist_horizontal = math.sqrt(x**2 + y**2)
#     vertical_angle = math.atan2(z, dist_horizontal)

#     return horizontal_angle, vertical_angle
