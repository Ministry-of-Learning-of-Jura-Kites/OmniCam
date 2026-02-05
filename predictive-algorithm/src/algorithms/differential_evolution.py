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
from state import State, render_from_state
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
    num_faces = len(template.faces)
    num_cams = len(template.cameras)

    bounds = []
    for _ in range(num_cams):
        bounds.extend(
            [
                (1.0, 10.0),  # Radius
                (-180.0, 180.0),  # Theta (Azimuth)
                (-45.0, 45.0),  # Phi (Elevation)
                (0, num_faces - 1),  # Face Index
            ]
        )

    num_particles = 20
    dim = len(initial_vec)
    init_pop = np.zeros((num_particles, dim))

    for i in range(num_particles):
        particle = initial_vec.copy()

        # 2. Strategic Diversification
        if i >= num_particles // 2:
            for cam_idx in range(num_cams):
                # Flip Azimuth
                particle[cam_idx * 4 + 1] = (particle[cam_idx * 4 + 1] + 180) % 360
                if particle[cam_idx * 4 + 1] > 180:
                    particle[cam_idx * 4 + 1] -= 360

                # Jiggle the Face Index: Half the population jumps to a random face
                particle[cam_idx * 4 + 3] = np.random.randint(0, num_faces)

        # 3. Add Noise
        # Noise for [r, theta, phi]
        spatial_noise = np.random.uniform(-5, 5, dim)

        # Zero out noise for face_idx to prevent float-rounding issues in initial seeds
        # We want the seeds to start exactly on faces, not between them.
        for cam_idx in range(num_cams):
            spatial_noise[cam_idx * 4 + 3] = 0

        init_pop[i] = particle + spatial_noise

        # Final safety clip to respect bounds
        lower_b = [b[0] for b in bounds]
        upper_b = [b[1] for b in bounds]
        init_pop[i] = np.clip(init_pop[i], lower_b, upper_b)

    def objective(vec):
        state = spherical_vector_to_state(vec, template)
        cost = total_cost(state, True)
        render_from_state(None, state)
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
        rng=200,
    )
    return spherical_vector_to_state(result.x, template)
