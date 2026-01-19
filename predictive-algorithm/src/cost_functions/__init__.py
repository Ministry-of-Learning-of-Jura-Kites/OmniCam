import logging
from state import State
from . import angle_cost, resolution_cost, occlusion_cost, mounting_cost


def total_cost(state: State):
    angle = angle_cost.cost(state)
    resolution = resolution_cost.cost(state)
    occlusion = occlusion_cost.cost(state)
    mounting = mounting_cost.cost(state)
    return angle + resolution + occlusion + mounting, {
        "angle": angle,
        "res": resolution,
        "occ": occlusion,
        "mount": mounting,
    }
