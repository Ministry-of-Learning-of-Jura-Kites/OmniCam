from typing import Union
import numpy as np
import quaternion
from state import CameraState, State
from constant import BIG_M
from basic_types import Array3


def is_in_view(point, cam_state: CameraState) -> Union[bool, Union[Array3]]:
    # 1. Transform point to Camera Local Space
    rel_point = point - cam_state.pos
    local_point = quaternion.rotate_vectors(cam_state.angle.conj(), rel_point)

    # Map variables to axes: X=Forward, Y=Up, Z=Horizontal
    x, y, z = local_point

    # 2. Depth Check (X is Forward)
    if x <= 0:
        return (False, local_point)

    # 3. Frustum Math
    aspect_ratio = cam_state.pixels[0] / cam_state.pixels[1]
    tan_half_vfov = np.tan(np.deg2rad(cam_state.vfov) / 2)

    depth = x

    # limit_y (Vertical) is determined by VFOV
    limit_y = depth * tan_half_vfov

    # limit_z (Horizontal) is vertical limit scaled by aspect ratio
    limit_z = limit_y * aspect_ratio

    # 4. NDC Validation
    # Vertical check (y-axis) and Horizontal check (z-axis)
    is_visible = (-limit_y <= y <= limit_y) and (-limit_z <= z <= limit_z)

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
