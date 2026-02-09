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

    num_faces = len(template.faces)
    num_cams = len(template.cameras)

    # cartesian = CartesianSerialize()
    spherical = SphericalSerialize(num_cams, num_faces, template.scale)

    # initial_vec = cartesian.state_to_vector(initial_state)
    initial_vec = spherical.state_to_vector(initial_state)

    # bounds = cartesian.init_bounds(template)
    bounds = spherical.init_bounds()

    num_particles = 20
    # init_pop = cartesian.init_pop(num_particles, initial_vec, bounds, num_cams)
    init_pop = spherical.init_pop(
        num_particles, initial_vec, bounds, num_cams, num_faces
    )

    def objective(vec):
        # state = cartesian.vector_to_state(vec, template)
        state = spherical.vector_to_state(vec, template)
        cost = total_cost(state, True)
        render_from_state(None, state)
        return cost

    result = differential_evolution(
        objective,
        bounds,
        strategy="rand1bin",
        maxiter=50,
        init=init_pop,
        mutation=(0.2, 0.7),
        popsize=100,
        recombination=0.7,
        rng=20000,
    )
    # return cartesian.vector_to_state(result.x, template)
    return spherical.vector_to_state(result.x, template)
