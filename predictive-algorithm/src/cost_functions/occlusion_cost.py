from typing import Union
import numpy as np
import quaternion
from state import CameraState, State
from constant import BIG_M
from basic_types import Array3


def is_in_view(point, cam_state: CameraState) -> Union[bool, Union[Array3]]:
    # Transform point to Camera Local Space (View Space)
    rel_point = point - cam_state.pos
    # Rotate (using the conjugate to move from world to local)
    local_point = quaternion.rotate_vectors(cam_state.angle.conj(), rel_point)

    x, y, z = local_point
    if y <= 0:
        return (False, local_point)

    aspect_ratio = cam_state.pixels[0] / cam_state.pixels[1]
    tan_half_vfov = np.tan(np.deg2rad(cam_state.vfov) / 2)

    # depth is the distance along the forward axis (y)
    depth = y

    # limit_z is the max vertical distance from the center at this depth
    limit_z = depth * tan_half_vfov
    # limit_x is the max horizontal distance from the center at this depth
    limit_x = limit_z * aspect_ratio

    # NDC Validation
    is_visible = (-limit_x <= x <= limit_x) and (-limit_z <= z <= limit_z)

    return (is_visible, local_point)


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        for corner in cam_state.face:
            valid_coord, ndc_coord = is_in_view(corner, cam_state)
            if not valid_coord:
                cost += BIG_M + np.linalg.norm(ndc_coord) * 100
                break

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
