from pyvistaqt import BackgroundPlotter
from scipy.optimize import differential_evolution
from . import state_to_vector, state_vector_dim, vector_to_state
import numpy as np
from state import State
from cost_functions import total_cost


def optimize_de(initial_state: State):
    template = initial_state
    dim = len(template.cameras) * state_vector_dim()

    # Define bounds
    bounds = []
    for _ in range(len(template.cameras)):
        bounds.extend(
            [(-100 / initial_state.scale, 100 / initial_state.scale)] * 3
        )  # Pos
        # bounds.extend([(-1, 1)] * 4)  # Quaternion components

    def objective(vec):
        cost, _ = total_cost(vector_to_state(vec, template))
        return cost

    result = differential_evolution(objective, bounds, strategy="best1bin", maxiter=50)
    return vector_to_state(result.x, template)
