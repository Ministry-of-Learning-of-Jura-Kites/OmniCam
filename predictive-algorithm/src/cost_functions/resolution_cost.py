import math
import numpy as np
import quaternion
from utils import center_of_face
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

    # print("x_pixel_per_dist", x_pixel_per_dist, "y_pixel_per_dist: ", y_pixel_per_dist)

    return x_pixel_per_dist, y_pixel_per_dist


def get_distance_to_face(face: Array4x3, cam_pos: Array3):
    face_center = center_of_face(face)

    return np.linalg.norm(face_center - cam_pos)


def ppm_to_cost(ppd: float):
    target_ppd = 120  # The "ideal" resolution
    min_acceptable = 80
    if ppd < min_acceptable:
        return 1000 + (min_acceptable - ppd) ** 2

    diff = max(0, target_ppd - ppd)
    cost = 500 * (1 - np.exp(-0.03 * diff))

    return max(cost, 0.0001)


def cost(state: State):
    cost = 0
    min_dist_threshold = 1 * state.scale
    for cam_state in state.cameras:
        face_dist = get_distance_to_face(cam_state.face, cam_state.pos)

        if face_dist < min_dist_threshold:
            # Quadratic penalty: The closer it gets to 0, the higher the cost explodes
            # At face_dist = min_dist_threshold, this is 0.
            # At face_dist = 0, this is 10,000.
            dist_error = min_dist_threshold - face_dist
            cost += 1000 * (dist_error / min_dist_threshold) ** 2

        x_pixel_per_dist, y_pixel_per_dist = get_pixel_per_meter(cam_state, face_dist)
        cost += ppm_to_cost(x_pixel_per_dist * state.scale)
        cost += ppm_to_cost(y_pixel_per_dist * state.scale)

    # print("resolution cost: ", cost)

    return cost
