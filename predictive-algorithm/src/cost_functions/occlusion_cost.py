from typing import Union
import numpy as np
import quaternion
from state import CameraState, State
from constant import BIG_M
from basic_types import Array3


def is_in_view(point, cam_state: CameraState) -> Union[bool, Union[Array3]]:
    # 1. Transform point to Camera Local Space (View Space)
    # Translate
    rel_point = point - cam_state.pos
    # Rotate (using the conjugate to move from world to local)
    local_point = quaternion.rotate_vectors(cam_state.angle.conj(), rel_point)

    # In standard camera systems, the camera looks down -Z or +Z.
    # Assuming standard GL convention: Camera looks down -Z
    x, y, z = local_point

    if z <= 0:  # Point is behind the camera
        return (False, local_point)

    # 2. Calculate Projection limits
    # aspect_ratio = width / height
    aspect_ratio = cam_state.pixels[0] / cam_state.pixels[1]
    tan_half_vfov = np.tan(np.deg2rad(cam_state.vfov) / 2)

    # 3. NDC Validation
    # The maximum allowed Y at this distance (z) is: |z| * tan(vfov/2)
    # The maximum allowed X at this distance (z) is: |z| * tan(vfov/2) * aspect_ratio

    limit_y = abs(z) * tan_half_vfov
    limit_x = limit_y * aspect_ratio

    return ((-limit_x <= x <= limit_x) and (-limit_y <= y <= limit_y), local_point)


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        for corner in cam_state.face:
            valid_coord, ndc_coord = is_in_view(corner, cam_state)
            if not valid_coord:
                cost += BIG_M + np.linalg.norm(ndc_coord) * 100
                continue

            point, _ = state.gltf.ray_trace(
                origin=cam_state.pos, end_point=corner, first_point=True
            )

            if len(point) == 0:
                continue
            to_corner_distance = np.linalg.norm(corner - cam_state.pos)
            to_intersected_distance = np.linalg.norm(point - cam_state.pos)
            distance = to_corner_distance - to_intersected_distance
            if to_corner_distance <= to_intersected_distance:
                continue
            elif distance < 0.2:
                cost += distance * 10
            else:
                cost += distance**4 * 10 + 0.2 * 10 - 0.2**4 * 10
    print("occlusion cost: ", cost)
    return cost
