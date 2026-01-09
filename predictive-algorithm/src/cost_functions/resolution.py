import math
import numpy as np
from state import CameraState, State
from basic_types import Array4x3, Array3


def get_pixel_per_distance(cam_state: CameraState, distance_to_plane: float):
    """Returns pixels/virtual metre unit"""
    fov_rad = math.radians(cam_state.fov)

    # Calculate the total physical width of the view at distance_to_plane
    # Width = 2 * d * tan(theta / 2)
    total_width = 2 * distance_to_plane * math.tan(fov_rad / 2)

    # Calculate height based on pixel aspect ratio
    aspect_ratio = cam_state.pixels[1] / cam_state.pixels[0]
    total_height = total_width * aspect_ratio

    # Distance represented by a single pixel
    x_pixel_per_dist = cam_state.pixels[0] / total_width
    y_pixel_per_dist = cam_state.pixels[1] / total_height

    return x_pixel_per_dist, y_pixel_per_dist


def get_distance_to_face(cam_pos: Array3, face: Array4x3):
    # face is Array4x3 -> take first 3 vertices
    v0, v1, v2 = face[0], face[1], face[2]

    # Calculate the normal vector of the face
    vec_a = v1 - v0
    vec_b = v2 - v0
    normal = np.cross(vec_a, vec_b)

    # Calculate the distance from camera pos to the plane
    # Dot product of (pos - v0) and the normal
    numerator = np.abs(np.dot(normal, (cam_pos - v0)))
    denominator = np.linalg.norm(normal)

    return numerator / denominator


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        face_dist = get_distance_to_face(cam_state.pos, cam_state.face)
        x_pixel_per_dist, y_pixel_per_dist = get_pixel_per_distance(
            cam_state, face_dist
        )
        cost += x_pixel_per_dist * state.scale
        cost += y_pixel_per_dist * state.scale

    return cost
