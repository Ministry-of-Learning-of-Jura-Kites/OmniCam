import math
import random
from typing import List, Tuple
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


def center_of_faces(faces: List[Array4x3]) -> Array3:
    if not faces:
        return np.array([0.0, 0.0, 0.0])

    # Calculate the center of each individual face
    individual_centers = [np.mean(face, axis=0) for face in faces]

    # Average of all those centers
    return np.mean(individual_centers, axis=0)


def normal_vec_of_face(face: Array4x3) -> Array3:
    A = face[0]
    B = face[1]
    C = face[2]

    v1 = B - A
    v2 = C - A

    normal = np.cross(v1, v2)

    norm = np.linalg.norm(normal)

    if norm < 1e-12:
        # Handle degenerate faces (where points are collinear)
        return np.zeros(3)

    return normal / norm


def angle_from_face_normal(
    face: Array4x3,
    pos: Array3,
    angle: quaternion.quaternion,
) -> Tuple[Quantity[u.radian], Quantity[u.radian]]:

    # 1. Standard vectors
    angle_vec = quaternion.rotate_vectors(angle, [1, 0, 0])
    face_center = np.mean(face, axis=0)
    face_normal = normal_vec_of_face(face)

    # 2. DECIDE: Front or Back?
    # Vector from face to camera
    face_to_cam = pos - face_center

    # Dot product: > 0 means camera is in front, < 0 means behind
    is_behind = np.dot(face_to_cam, face_normal) < 0

    # If behind, we want to treat the "back" as the target normal
    target_normal = -face_normal if is_behind else face_normal

    # 3. Horizontal Offset (XZ Plane)
    # Use target_normal instead of face_normal
    horiz_look_angle = math.atan2(-target_normal[2], -target_normal[0])
    horiz_cam_angle = math.atan2(angle_vec[2], angle_vec[0])
    horizontal_offset = horiz_look_angle - horiz_cam_angle

    # 4. Vertical Offset
    dist_xz_look = math.sqrt(target_normal[0] ** 2 + target_normal[2] ** 2)
    dist_xz_cam = math.sqrt(angle_vec[0] ** 2 + angle_vec[2] ** 2)

    vert_look_angle = math.atan2(-target_normal[1], dist_xz_look)
    vert_cam_angle = math.atan2(angle_vec[1], dist_xz_cam)
    vertical_offset = vert_look_angle - vert_cam_angle

    # 5. Normalize
    horizontal_offset = (horizontal_offset + math.pi) % (2 * math.pi) - math.pi

    # 6. Add a "Side Penalty" (Optional)
    # If you prefer the front over the back, add a small flat cost if is_behind == True

    return (
        horizontal_offset * u.radian,
        vertical_offset * u.radian,
    )


def look_at_quaternion(forward_vector, up_vector=np.array([0, 1, 0])):
    # 1. Normalize the forward vector (Z-axis in many systems, but we'll use your X-forward)
    # We want to map our local Forward (1,0,0) to target_forward
    f = forward_vector / (np.linalg.norm(forward_vector) + 1e-8)

    # 2. Calculate the Right vector
    # World Up cross Forward = Right
    r = np.cross(f, up_vector)
    if np.linalg.norm(r) < 1e-6:
        # Fallback if looking straight up or down
        r = np.array([0, 0, 1])
    r /= np.linalg.norm(r)

    # 3. Calculate the true local Up vector
    # Forward cross Right = Up
    u = np.cross(r, f)

    # 4. Construct Rotation Matrix
    # If your camera's local forward is +X and up is +Y:
    # Column 0: Forward (X)
    # Column 1: Up (Y)
    # Column 2: Right (Z)
    rot_matrix = np.stack([f, u, r], axis=1)

    # 5. Convert Matrix to Quaternion
    return quaternion.from_rotation_matrix(rot_matrix)
