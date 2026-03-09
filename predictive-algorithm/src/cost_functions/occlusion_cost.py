from typing import Union
import numpy as np
import quaternion
from state import CameraState, State
from basic_types import Array3, Array4x3
from utils import center_of_face
import vtk

import numpy as np


def generate_raycast_depth_map(state, cam_state, resolution=(100, 100)):
    """
    Generates a depth map via ray casting.
    Result is a 2D array where each value is the distance from the camera.
    """
    depth_map = np.zeros(resolution)
    h_res, v_res = resolution

    # Extract camera parameters
    pos = cam_state.pos
    # Create normalized direction vectors for the camera
    # We assume 'forward', 'up', and 'right' vectors are derived from cam_state.angle
    forward = np.array([0, 0, -1])
    up = np.array([0, 1, 0])
    right = forward.cross(up)

    # Calculate FOV-based focal length
    fov_rad = np.deg2rad(cam_state.camera_config.vfov)
    focal_length = 1.0 / np.tan(fov_rad / 2)

    for v in range(v_res):
        for u in range(h_res):
            # Map pixel (u,v) to [-1, 1] range
            ndc_u = 2 * u / h_res - 1
            ndc_v = (2 * v / v_res - 1) * (v_res / h_res)  # Adjust for aspect

            # Compute ray direction
            ray_dir = forward * focal_length + right * ndc_u + up * ndc_v
            ray_dir /= np.linalg.norm(ray_dir)

            # Use the same VTK locator as your optimization code
            points = vtk.vtkPoints()
            cellIds = vtk.vtkIdList()

            # Cast ray
            hit = state.gltf_locator.IntersectWithLine(
                pos, pos + ray_dir * 1000, 0.001, points, cellIds
            )

            if hit:
                # Store distance to intersection
                hit_point = np.array(points.GetPoint(0))
                depth_map[v, u] = np.linalg.norm(hit_point - pos)
            else:
                # Far plane default
                depth_map[v, u] = 1000.0

    return depth_map


def is_in_view(point, cam_state: CameraState) -> Union[bool, Union[Array3]]:
    # Transform point to Camera Local Space
    rel_point = point - cam_state.pos
    local_point = quaternion.rotate_vectors(cam_state.angle.conj(), rel_point)

    # Map variables to axes: X=Forward, Y=Up, Z=Horizontal
    x, y, z = local_point

    depth = -z

    # Depth Check (X is Forward)
    if depth <= 0:
        return (False, local_point)

    aspect_ratio = cam_state.camera_config.pixels[0] / cam_state.camera_config.pixels[1]
    tan_half_vfov = np.tan(np.deg2rad(cam_state.camera_config.vfov) / 2)

    # limit_y (Vertical) is determined by VFOV
    limit_y = depth * tan_half_vfov

    # limit_z (Horizontal) is vertical limit scaled by aspect ratio
    limit_x = limit_y * aspect_ratio

    # NDC Validation
    is_visible = (-limit_y <= y <= limit_y) and (-limit_x <= x <= limit_x)

    return (is_visible, local_point)


def cost_single_cam(state: State, cam_state: CameraState, face: Array4x3):
    total_occlusion_cost = 0

    face_center = center_of_face(face)
    check_corners = [
        (corner * (100 - i * 10) / 100 + face_center * i * 10 / 100, i)
        for corner in face
        for i in range(0, 3)
    ]
    for corner, weight in check_corners:
        # 1. Soften the 'Out of View' penalty
        valid_coord, ndc_coord = is_in_view(corner, cam_state)
        if not valid_coord:
            # Instead of BIG_M, use the distance to the screen edge.
            # ndc_coord usually ranges from -1 to 1.
            # If it's 1.5, we want to guide it back to 1.0.
            dist_outside = np.max(np.abs(ndc_coord) - 1.0, initial=0)
            total_occlusion_cost += 500 * (weight + 1) + (dist_outside * 100)
            continue  # If not in view, occlusion check is secondary

        points = vtk.vtkPoints()  # Stores the intersection coordinates
        cellIds = vtk.vtkIdList()
        tolerance = 0.1
        code = state.gltf_locator.IntersectWithLine(
            cam_state.pos, corner, tolerance, points, cellIds
        )

        if code == 0:
            continue

        to_corner_dist = np.linalg.norm(corner - cam_state.pos)
        to_hit_dist = np.linalg.norm(points.GetPoint(0) - cam_state.pos)

        # distance is how much of the ray is 'blocked'
        # We only care if hit_dist < corner_dist
        if to_hit_dist < to_corner_dist:
            blocked_depth = to_corner_dist - to_hit_dist

            # Use a Quadratic Penalty instead of d**4
            # It's steep enough to be a 'hard' constraint,
            # but numerically more stable.
            total_occlusion_cost += 1000 * (blocked_depth**2)
    return total_occlusion_cost


def cost(state: State):
    total_occlusion_cost = 0

    for cam_state in state.cameras:
        total_occlusion_cost += cost_single_cam(state, cam_state)
    return total_occlusion_cost
