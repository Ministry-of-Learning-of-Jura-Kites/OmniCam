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


# TODO: Correct horizontal angle
def angle_from_face_normal(
    face: Array4x3,
    pos: Array3,
    angle: quaternion.quaternion,
) -> Tuple[Quantity[u.radian], Quantity[u.radian]]:
    """Returns vertical and horizontal angles in radian unit"""

    angle_vec = quaternion.rotate_vectors(angle, [0, 1, 0])

    # Define Highest point of the face in Y as the target
    face_center = np.mean(face, axis=0)
    highest_face_center = np.array([face_center[0], face[:, 1].max(), face_center[2]])
    look_vec = highest_face_center - pos

    # Horizontal Offset (XZ Plane)
    horiz_look_angle = math.atan2(look_vec[2], look_vec[0])
    horiz_cam_angle = math.atan2(angle_vec[2], angle_vec[0])
    horizontal_offset = horiz_look_angle - horiz_cam_angle

    # Vertical Offset (Y is Up)
    # Magnitude in the ground plane (XZ)
    dist_xz_look = math.sqrt(look_vec[0] ** 2 + look_vec[2] ** 2)
    dist_xz_cam = math.sqrt(angle_vec[0] ** 2 + angle_vec[2] ** 2)

    # Angle from the ground plane up toward the Y axis
    vert_look_angle = math.atan2(look_vec[1], dist_xz_look)
    vert_cam_angle = math.atan2(angle_vec[1], dist_xz_cam)

    vertical_offset = vert_look_angle - vert_cam_angle

    return (
        horizontal_offset * u.radian,
        vertical_offset * u.radian,
    )
