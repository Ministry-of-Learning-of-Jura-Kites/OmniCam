import math
import numpy as np
from state import CameraState, State
from basic_types import Array4x3, Array3
from constant import BIG_M


def get_pixel_per_meter(cam_state: CameraState, distance_to_plane: float):
    """Returns pixels/virtual metre unit"""
    vfov_rad = math.radians(cam_state.vfov)

    total_height = 2 * distance_to_plane * math.tan(vfov_rad / 2)

    # 3. Calculate width based on the aspect ratio (Width / Height)
    # Aspect Ratio = W_px / H_px
    aspect_ratio = cam_state.pixels[0] / cam_state.pixels[1]
    total_width = total_height * aspect_ratio

    # Distance represented by a single pixel
    x_pixel_per_dist = cam_state.pixels[0] / total_width
    y_pixel_per_dist = cam_state.pixels[1] / total_height

    print("x_pixel_per_dist", x_pixel_per_dist, "y_pixel_per_dist: ", y_pixel_per_dist)

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


def ppm_to_cost(ppd: float):
    base = 3
    cost_at_20 = 200
    if ppd < 20:
        return BIG_M - ppd * BIG_M / 20 + cost_at_20
    elif ppd < 200:
        base = (cost_at_20 + 1) ** (1 / (200 - 20))
        return base ** (-(ppd - 200)) - 1
    else:
        return 0


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        face_dist = get_distance_to_face(cam_state.pos, cam_state.face)
        x_pixel_per_dist, y_pixel_per_dist = get_pixel_per_meter(cam_state, face_dist)
        cost += ppm_to_cost(x_pixel_per_dist * state.scale)
        cost += ppm_to_cost(y_pixel_per_dist * state.scale)

    print("resolution cost: ", cost)

    return cost
