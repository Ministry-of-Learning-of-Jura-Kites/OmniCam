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
    total_occlusion_cost = 0

    for cam_state in state.cameras:
        for corner in cam_state.face:
            # 1. Soften the 'Out of View' penalty
            valid_coord, ndc_coord = is_in_view(corner, cam_state)
            if not valid_coord:
                # Instead of BIG_M, use the distance to the screen edge.
                # ndc_coord usually ranges from -1 to 1.
                # If it's 1.5, we want to guide it back to 1.0.
                dist_outside = np.max(np.abs(ndc_coord) - 1.0, initial=0)
                total_occlusion_cost += 500 + (dist_outside * 100)
                continue  # If not in view, occlusion check is secondary

            # 2. Ray-trace check
            point, _ = state.gltf.ray_trace(
                origin=cam_state.pos, end_point=corner, first_point=True
            )

            if len(point) == 0:
                continue

            to_corner_dist = np.linalg.norm(corner - cam_state.pos)
            to_hit_dist = np.linalg.norm(point - cam_state.pos)

            # distance is how much of the ray is 'blocked'
            # We only care if hit_dist < corner_dist
            if to_hit_dist < to_corner_dist:
                blocked_depth = to_corner_dist - to_hit_dist

                # Use a Quadratic Penalty instead of d**4
                # It's steep enough to be a 'hard' constraint,
                # but numerically more stable.
                total_occlusion_cost += 1000 * (blocked_depth**2)

    return total_occlusion_cost
