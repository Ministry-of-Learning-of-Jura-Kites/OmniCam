from state import State
from . import angle_cost, resolution_cost, occlusion_cost


def total_cost(state: State, log: bool = False):
    angle = angle_cost.cost(state)
    resolution = 50 * resolution_cost.cost(state)
    occlusion = occlusion_cost.cost(state)
    if log:
        print("angle cost: ", angle)
        print("resolution cost: ", resolution)
        print("occlusion cost: ", occlusion)
    return angle + resolution + occlusion
