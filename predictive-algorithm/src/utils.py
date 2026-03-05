import math
import random
from typing import List, Tuple
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


def center_of_faces(faces: List[Array4x3]) -> Array3:
    # if not faces:
    #     return np.array([0.0, 0.0, 0.0])

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

    # 1. Standard vectors: Rotate the LOCAL -Z vector
    angle_vec = quaternion.rotate_vectors(angle, [0, 0, -1])
    face_center = np.mean(face, axis=0)
    face_normal = normal_vec_of_face(face)

    # 2. DECIDE: Front or Back?
    face_to_cam = pos - face_center
    is_behind = np.dot(face_to_cam, face_normal) < 0
    target_normal = -face_normal if is_behind else face_normal

    # 3. Horizontal Offset (Now in the XZ plane, with Z as the "forward" baseline)
    # We want to know the angle on the ground plane (X, Z)
    # Target direction to look is -target_normal
    target_dir = -target_normal

    # In -Z forward, atan2(x, -z) gives the angle relative to the forward vector
    horiz_look_angle = math.atan2(target_dir[0], -target_dir[2])
    horiz_cam_angle = math.atan2(angle_vec[0], -angle_vec[2])
    horizontal_offset = horiz_look_angle - horiz_cam_angle

    # 4. Vertical Offset (Pitch)
    # dist_xz is the "adjacent" side of the triangle for the pitch angle
    dist_xz_look = math.sqrt(target_dir[0] ** 2 + target_dir[2] ** 2)
    dist_xz_cam = math.sqrt(angle_vec[0] ** 2 + angle_vec[2] ** 2)

    # Pitch is atan2(y, distance_in_plane)
    vert_look_angle = math.atan2(target_dir[1], dist_xz_look)
    vert_cam_angle = math.atan2(angle_vec[1], dist_xz_cam)
    vertical_offset = vert_look_angle - vert_cam_angle

    # 5. Normalize Horizontal Offset to [-pi, pi]
    horizontal_offset = (horizontal_offset + math.pi) % (2 * math.pi) - math.pi

    return (
        horizontal_offset * u.radian,
        vertical_offset * u.radian,
    )


def angle_from_face_position(
    face: Array4x3,
    pos: Array3,
    angle: quaternion.quaternion,
) -> Tuple[Quantity[u.radian], Quantity[u.radian]]:

    # 1. Camera's actual forward direction
    # Now rotating the local -Z vector
    cam_forward_vec = quaternion.rotate_vectors(angle, [0, 0, -1])

    # 2. Vector from Camera to Face Center (The "Target" vector)
    face_center = np.mean(face, axis=0)
    target_vec = face_center - pos

    # Normalize the target vector
    target_dist = np.linalg.norm(target_vec)
    target_unit = target_vec / (target_dist + 1e-8)

    # 3. Horizontal Offset (XZ Plane)
    # Using atan2(x, -z) makes the -Z axis (forward) 0 radians.
    # Positive x results in a positive angle (turning right).
    horiz_target_angle = math.atan2(target_unit[0], -target_unit[2])
    horiz_cam_angle = math.atan2(cam_forward_vec[0], -cam_forward_vec[2])
    horizontal_offset = horiz_target_angle - horiz_cam_angle

    # 4. Vertical Offset (Pitch)
    # Distance in the XZ plane acts as the adjacent side
    dist_xz_target = math.sqrt(target_unit[0] ** 2 + target_unit[2] ** 2)
    dist_xz_cam = math.sqrt(cam_forward_vec[0] ** 2 + cam_forward_vec[2] ** 2)

    # Elevation is atan2(y, distance_in_plane)
    vert_target_angle = math.atan2(target_unit[1], dist_xz_target)
    vert_cam_angle = math.atan2(cam_forward_vec[1], dist_xz_cam)
    vertical_offset = vert_target_angle - vert_cam_angle

    # 5. Normalize Horizontal to [-pi, pi]
    horizontal_offset = (horizontal_offset + math.pi) % (2 * math.pi) - math.pi

    return (
        horizontal_offset * u.radian,
        vertical_offset * u.radian,
    )


def look_at_quaternion(forward_vector, up_vector=np.array([0, 1, 0])):
    # 1. Normalize the target direction
    # This is the direction we want the LOCAL -Z to point towards
    f_target = forward_vector / (np.linalg.norm(forward_vector) + 1e-8)

    # 2. Calculate the Right vector (+X)
    # In RHR: Right = Forward x Up.
    # Since our target is -Z, we use (Target Forward) x (Up) to get Right
    r = np.cross(f_target, up_vector)

    if np.linalg.norm(r) < 1e-6:
        # Fallback for gimbal lock (looking straight up/down)
        r = np.array([1, 0, 0])
    r /= np.linalg.norm(r)

    # 3. Calculate the true local Up vector (+Y)
    # Up = Right x Forward
    u = np.cross(r, f_target)

    # 4. Construct Rotation Matrix
    # Column 0: Right (+X)
    # Column 1: Up (+Y)
    # Column 2: BACKWARD (+Z)
    # Because our forward is -Z, the matrix needs the backward vector ( -f_target )
    rot_matrix = np.stack([r, u, -f_target], axis=1)

    # 5. Convert Matrix to Quaternion
    return quaternion.from_rotation_matrix(rot_matrix)
