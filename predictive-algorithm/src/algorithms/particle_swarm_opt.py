from typing import Union
from pyvistaqt import BackgroundPlotter
from . import state_to_vector, state_vector_dim, vector_to_state
import numpy as np
from state import InteractiveOptimizerPlotter, State
from cost_functions import total_cost


def optimize_pso(
    initial_state: State,
    pl: Union[BackgroundPlotter, None],
    iterations=100,
    pop_size=30,
):
    viz: Union[InteractiveOptimizerPlotter, None] = None
    if pl != None:
        viz = InteractiveOptimizerPlotter(pl, initial_state)

    num_cams = len(initial_state.cameras)
    dim = num_cams * state_vector_dim()

    # Initialize particles
    particles = np.array([state_to_vector(initial_state) for _ in range(pop_size)])
    # Add some noise to initial particles to explore
    particles += np.random.uniform(-0.5, 0.5, particles.shape)

    velocities = np.zeros_like(particles)
    p_best = particles.copy()
    p_best_cost = np.array(
        [total_cost(vector_to_state(p, initial_state))[0] for p in particles]
    )

    g_best = p_best[np.argmin(p_best_cost)]
    g_best_cost = np.min(p_best_cost)

    # Hyperparameters
    w, c1, c2 = 0.5, 1.5, 1.5

    for iteration in range(iterations):
        for i in range(pop_size):
            # Update velocity and position
            r1, r2 = np.random.rand(dim), np.random.rand(dim)
            velocities[i] = (
                w * velocities[i]
                + c1 * r1 * (p_best[i] - particles[i])
                + c2 * r2 * (g_best - particles[i])
            )
            particles[i] += velocities[i]

            # Evaluate
            current_state = vector_to_state(particles[i], initial_state)
            current_cost = total_cost(current_state)

            if current_cost < p_best_cost[i]:
                p_best[i] = particles[i]
                p_best_cost[i] = current_cost

                if current_cost < g_best_cost:
                    g_best = particles[i]
                    g_best_cost = current_cost
        if pl != None:
            best_state_so_far = vector_to_state(g_best, initial_state)
            print("Best state cost: ", total_cost(best_state_so_far))
            viz.update(best_state_so_far, iteration)

    return vector_to_state(g_best, initial_state)
