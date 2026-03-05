import math
from state import CameraState, State
import vtk
from basic_types import Array4x3


def cost_single_cam(state: State, cam_state: CameraState, _face: Array4x3):
    # Pre-allocate containers to avoid memory churn
    closest_p = [0.0, 0.0, 0.0]
    cell_id = vtk.reference(0)
    sub_id = vtk.reference(0)
    dist2 = vtk.reference(0.0)

    # 1. Let C++ do the heavy lifting (finding the absolute closest point)
    state.gltf_locator.FindClosestPoint(
        cam_state.pos, closest_p, cell_id, sub_id, dist2
    )

    d_val = math.sqrt(dist2.get())
    mounting_tolerance = 0.1 * state.scale

    # 2. Basic distance penalty (Quadratic "pull" toward surfaces)
    cost = 0.0
    if d_val > mounting_tolerance:
        cost += 5000 * ((d_val - mounting_tolerance) ** 2)

    # 3. The "Below" Penalty (Hard Push)
    # If the surface we found is lower than the camera, it's a floor/ledge.
    # We apply a heavy penalty to discourage the optimizer from staying here.
    if closest_p[1] < cam_state.pos[1] - 0.001:
        # We use a large constant + a multiplier to create a gradient
        # that leads the camera AWAY from the floor.
        cost += 20000 + (d_val * 1000)

    return cost


def cost(state: State):
    for cam_state in state.cameras:
        cost_single_cam(state, cam_state)
    return cost
