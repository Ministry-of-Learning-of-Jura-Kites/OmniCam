from scipy.optimize import differential_evolution
from . import state_to_vector, state_vector_dim, vector_to_state
import numpy as np
from state import State
from cost_functions import total_cost


def optimize_camera_de(initial_state: State):
    template = initial_state
    dim = len(template.cameras) * state_vector_dim()

    # Define bounds (example: pos +/- 10m, vfov 10 to 120 degrees)
    bounds = []
    for _ in range(len(template.cameras)):
        bounds.extend([(-10, 10)] * 3)  # Pos
        bounds.extend([(-1, 1)] * 4)  # Quaternion components

    def objective(vec):
        return total_cost(vector_to_state(vec, template))

    result = differential_evolution(objective, bounds, strategy="best1bin", maxiter=50)
    return vector_to_state(result.x, template)
