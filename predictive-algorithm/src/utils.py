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


def look_at_quaternion(
    forward_vector, up_vector=np.array([0, 1, 0]), reference_forward=np.array([0, 0, 1])
):
    """
    Generates a quaternion that rotates an object to look at a specific direction.

    Args:
        direction_vector (np.ndarray): The direction the object should face (e.g., target position - current position).
        up_vector (np.ndarray): The world's "up" direction.

    Returns:
        np.ndarray: A quaternion in scalar-last format (x, y, z, w).
    """
    # Normalize vectors
    target_forward = forward_vector / np.linalg.norm(forward_vector)
    global_up = up_vector / np.linalg.norm(up_vector)

    # Calculate the rotation axis (cross product)
    # This might be tricky if the vectors are parallel/anti-parallel.
    axis = np.cross(reference_forward, target_forward)

    # Check for edge cases where vectors are parallel (cross product is zero)
    if np.linalg.norm(axis) < 1e-6:
        # If parallel, no rotation needed.
        if np.dot(reference_forward, target_forward) > 0:
            return np.quaternion(1, 0, 0, 0)  # No rotation
        else:
            # If anti-parallel, rotate 180 degrees around an arbitrary axis perpendicular to forward.
            # Using the global up vector as a reference for a valid axis.
            axis = np.cross(reference_forward, global_up)
            if (
                np.linalg.norm(axis) < 1e-6
            ):  # global_up was parallel to reference_forward
                axis = np.array([1, 0, 0])  # Use X axis as fallback
            axis = axis / np.linalg.norm(axis)
            angle = np.pi  # 180 degrees

    else:
        # Normalize the axis
        axis = axis / np.linalg.norm(axis)
        # Calculate the angle (dot product and arccos)
        cos_angle = np.dot(reference_forward, target_forward)
        cos_angle = np.clip(cos_angle, -1.0, 1.0)  # avoid numerical issues
        angle = np.arccos(cos_angle)

    # Create SciPy Rotation object and get quaternion
    quat = quaternion.from_rotation_vector(axis * angle)
    return quat
