from typing import Union
from pyvistaqt import BackgroundPlotter
import numpy as np
from state import State
from cost_functions import total_cost
from . import CartesianSerialize


def optimize_pso(
    initial_state: State,
    pl: Union[BackgroundPlotter, None],
    seed: int = 2000,
    iterations=500,
    pop_size=50,
):
    cartesian = CartesianSerialize(seed)
    num_cams = len(initial_state.cameras)
    initial_vec = cartesian.state_to_vector(initial_state)
    bounds = cartesian.init_bounds(initial_state)

    particles = cartesian.init_pop(pop_size, initial_vec, bounds, num_cams)
    particles += np.random.uniform(-0.5, 0.5, particles.shape)

    velocities = np.zeros_like(particles)
    p_best = particles.copy()
    p_best_cost = np.array(
        [total_cost(cartesian.vector_to_state(p, initial_state)) for p in particles]
    )

    g_best = p_best[np.argmin(p_best_cost)]
    g_best_cost = np.min(p_best_cost)

    # Hyperparameters
    w, c1, c2 = 0.5, 1.5, 1.5
    dim = particles.shape[1]  # Using shape for robustness

    # --- New Tracking Variables ---
    generations = 0
    patience = 50
    min_delta = 1e-2
    no_improvement_counter = 0
    # ------------------------------

    for iteration in range(iterations):
        generations += 1
        prev_g_best_cost = g_best_cost

        for i in range(pop_size):
            r1, r2 = np.random.rand(dim), np.random.rand(dim)
            velocities[i] = (
                w * velocities[i]
                + c1 * r1 * (p_best[i] - particles[i])
                + c2 * r2 * (g_best - particles[i])
            )
            particles[i] += velocities[i]

            current_state = cartesian.vector_to_state(particles[i], initial_state)
            current_cost = total_cost(current_state)

            if current_cost < p_best_cost[i]:
                p_best[i] = particles[i].copy()
                p_best_cost[i] = current_cost

                if current_cost < g_best_cost:
                    g_best = particles[i].copy()
                    g_best_cost = current_cost

        # --- Early Stopping Logic ---
        # Check if the improvement is significant
        if (prev_g_best_cost - g_best_cost) < min_delta:
            no_improvement_counter += 1
        else:
            no_improvement_counter = 0

        if no_improvement_counter >= patience:
            print(
                f"Early stopping at iteration {iteration} due to lack of improvement."
            )
            break
        # ----------------------------

    return cartesian.vector_to_state(g_best, initial_state), generations
