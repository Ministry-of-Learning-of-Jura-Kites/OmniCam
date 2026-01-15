from pyvistaqt import BackgroundPlotter
from scipy.optimize import differential_evolution
from . import (
    spherical_vector_to_state,
    state_to_spherical_vector,
    state_to_vector,
    state_vector_dim,
    vector_to_state,
)
import numpy as np
from state import State
from cost_functions import total_cost


def optimize_de(initial_state: State):
    template = initial_state

    # Define bounds
    # bounds = []
    # for _ in range(len(template.cameras)):
    #     bounds.extend(
    #         [(-100 / initial_state.scale, 100 / initial_state.scale)] * 3
    #     )  # Pos
    #     # bounds.extend([(-1, 1)] * 4)  # Quaternion components

    initial_vec = state_to_spherical_vector(initial_state)

    bounds = []
    for _ in range(len(template.cameras)):
        bounds.extend([(1.0, 10.0), (-180.0, 180.0), (-45.0, 45.0)])

    num_particles = 20  # or popsize * dim
    dim = len(initial_vec)
    init_pop = np.zeros((num_particles, dim))

    for i in range(num_particles):
        particle = initial_vec.copy()
        if i >= num_particles // 2:
            # Flip the Azimuth (theta) for the second half of the population
            # index 1 is theta in [r, theta, phi]
            for cam_idx in range(len(template.cameras)):
                particle[cam_idx * 3 + 1] += 180.0
                # Keep it within [-180, 180]
                if particle[cam_idx * 3 + 1] > 180:
                    particle[cam_idx * 3 + 1] -= 360

        # Add wider noise to help initial spread
        init_pop[i] = particle + np.random.uniform(-5, 5, dim)

    def objective(vec):
        # cost, _ = total_cost(vector_to_state(vec, template))
        # return cost
        state = spherical_vector_to_state(vec, template)
        cost, _ = total_cost(state)
        return cost

    result = differential_evolution(
        objective,
        bounds,
        strategy="rand1bin",
        maxiter=50,
        init=init_pop,
        mutation=(0.5, 1.0),
        popsize=20,
        recombination=0.7,
    )
    return spherical_vector_to_state(result.x, template)
