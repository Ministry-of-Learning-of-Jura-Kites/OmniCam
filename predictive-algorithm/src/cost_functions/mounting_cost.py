import math
import numpy as np
from traitlets import Float
from state import CameraState, State
import vtk
from basic_types import Array4x3


def cost_single_cam(state: State, cam_state: CameraState, face: Array4x3):
    total_mounting_score = 0
    # The 'ideal' distance from the lidar surface (e.g., thickness of a mount)
    mounting_tolerance = 0.1 * state.scale

    closest_point = [0.0, 0.0, 0.0]
    cell_id = vtk.reference(0)
    sub_id = vtk.reference(0)
    dist2 = vtk.reference(0.0)

    state.gltf_locator.FindClosestPoint(
        cam_state.pos, closest_point, cell_id, sub_id, dist2
    )

    dist_to_surface = math.sqrt(dist2)

    if dist_to_surface > mounting_tolerance:
        # Quadratic penalty for "floating" cameras.
        # This pulls the camera toward the walls/ceiling.
        diff = dist_to_surface - mounting_tolerance
        return 5000 * (diff**2)
    else:
        # Optional: small reward or 0 cost for being perfectly on surface
        return 0


def cost(state: State):
    for cam_state in state.cameras:
        cost_single_cam(state, cam_state)
    return cost
