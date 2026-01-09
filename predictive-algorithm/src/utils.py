import math
import random
from typing import Tuple
import numpy as np
import quaternion
from basic_types import Array4x3, Array3
from astropy.units import Quantity
import astropy.units as u


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


def angle_from_point(
    point: Array3, pos: Array3, angle: quaternion.quaternion
) -> Tuple[Quantity[u.radian], Quantity[u.radian]]:
    """Returns vertical and horizontal angles in radian unit"""
    to_point_world = point - pos

    if np.linalg.norm(to_point_world) == 0:
        return 0.0, 0.0

    # Rotate the world vector into LOCAL space
    to_point_local = quaternion.rotate_vectors(angle.conjugate(), to_point_world)

    # Assuming:
    # Z-axis = Forward
    # X-axis = Right
    # Y-axis = Up
    x, y, z = to_point_local

    # Horizontal angle (Azimuth): angle around the Up (Y) axis
    # Angle between Forward (Z) and the vector projected on ZX plane
    horizontal_angle = math.atan2(x, z)

    # Vertical angle (Elevation): angle above/below the ZX plane
    # distance in the horizontal plane
    dist_horizontal = math.sqrt(x**2 + z**2)
    vertical_angle = math.atan2(y, dist_horizontal)

    return horizontal_angle, vertical_angle
