from typing import Union
from pyvistaqt import BackgroundPlotter
import numpy as np
from state import State
from cost_functions import total_cost
from . import CartesianSerialize


def optimize_random_search(
    initial_state: State,
    pl: Union[BackgroundPlotter, None],
    seed: int = 2000,
    iterations=500,
    pop_size=50,
):
    cartesian = CartesianSerialize(seed)
    num_cams = len(initial_state.cameras)

    # FIX: Convert list of tuples to a NumPy array for slicing
    bounds = np.array(cartesian.init_bounds(initial_state))

    initial_vec = cartesian.state_to_vector(initial_state)
    g_best = initial_vec.copy()
    g_best_cost = total_cost(initial_state)

    generations = 0
    patience = 50
    min_delta = 1e-2
    no_improvement_counter = 0

    for iteration in range(iterations):
        generations += 1
        prev_g_best_cost = g_best_cost

        # Now this slicing will work perfectly
        samples = np.random.uniform(
            low=bounds[:, 0], high=bounds[:, 1], size=(pop_size, len(initial_vec))
        )

        costs = np.array(
            [total_cost(cartesian.vector_to_state(s, initial_state)) for s in samples]
        )

        current_best_idx = np.argmin(costs)
        if costs[current_best_idx] < g_best_cost:
            g_best_cost = costs[current_best_idx]
            g_best = samples[current_best_idx].copy()
            no_improvement_counter = 0
        else:
            if (prev_g_best_cost - g_best_cost) < min_delta:
                no_improvement_counter += 1

        if no_improvement_counter >= patience:
            print(f"Random search stopped at iteration {iteration}.")
            break

    return cartesian.vector_to_state(g_best, initial_state), generations
