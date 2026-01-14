import logging
from state import State
from . import angle_cost, resolution_cost, occlusion_cost


def total_cost(state: State):
    angle = angle_cost.cost(state)
    resolution = 50 * resolution_cost.cost(state)
    occlusion = occlusion_cost.cost(state)
    return angle + resolution + occlusion, {
        "angle": angle,
        "res": resolution,
        "occ": occlusion,
    }
