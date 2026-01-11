from state import State
from . import angle_cost, resolution_cost, occlusion_cost


def total_cost(state: State):
    return (
        angle_cost.cost(state)
        + resolution_cost.cost(state)
        + occlusion_cost.cost(state)
    )
