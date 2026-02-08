from pyvistaqt import BackgroundPlotter
from scipy.optimize import differential_evolution
from . import (
    CartesianSerialize,
    SphericalSerialize,
)
import numpy as np
from state import State, render_from_state
from cost_functions import total_cost


def optimize_de(initial_state: State):
    template = initial_state

    cartesian = CartesianSerialize()
    # spherical = SphericalSerialize()

    initial_vec = cartesian.state_to_vector(initial_state)
    # initial_vec = spherical.state_to_vector(initial_state)
    num_faces = len(template.faces)
    num_cams = len(template.cameras)

    bounds = cartesian.init_bounds(template)
    # bounds = spherical.init_bounds(num_cams, num_faces)

    num_particles = 20
    dim = len(initial_vec)
    init_pop = cartesian.init_pop(num_particles, initial_vec, bounds, num_cams)
    print(init_pop)
    # init_pop = spherical.init_pop(
    #     num_particles, initial_vec, bounds, num_cams, num_faces
    # )

    def objective(vec):
        state = cartesian.vector_to_state(vec, template)
        # state = spherical.vector_to_state(vec, template)
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
    return cartesian.vector_to_state(result.x, template)
    # return spherical.vector_to_state(result.x, template)
