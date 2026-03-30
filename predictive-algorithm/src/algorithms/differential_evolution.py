from scipy.optimize import differential_evolution
from . import CartesianSerialize
from state import State
from cost_functions import total_cost


def optimize_de(initial_state: State, seed: int, verbose=False):
    template = initial_state

    num_cams = len(template.cameras)

    cartesian = CartesianSerialize(seed)
    # spherical = SphericalSerialize(num_cams, num_faces, template.scale, seed)

    initial_vec = cartesian.state_to_vector(initial_state)
    # initial_vec = spherical.state_to_vector(initial_state)

    bounds = cartesian.init_bounds(template)
    # bounds = spherical.init_bounds()

    num_particles = 50
    init_pop = cartesian.init_pop(num_particles, initial_vec, bounds, num_cams)
    # init_pop = spherical.init_pop(
    #     num_particles, initial_vec, bounds, num_cams, num_faces
    # )

    def objective(vec):
        state = cartesian.vector_to_state(vec, template)
        # state = spherical.vector_to_state(vec, template)
        cost = total_cost(state, verbose)
        # render_from_state(None, state)
        return cost

    result = differential_evolution(
        objective,
        bounds,
        strategy="rand1bin",
        maxiter=500,
        init=init_pop,
        mutation=(0.2, 0.7),
        popsize=num_particles,
        recombination=0.9,
        rng=seed,
        # Fitness Stagnation:
        # tol=0 ignores relative change in favor of atol
        tol=0,
        # atol matches epsilon (10^-6)
        atol=1e-2,
        # Ensure it checks for 50 consecutive generations
        # Note: SciPy's internal 'convergence' check varies slightly by version,
        # but setting polish=False ensures it stops strictly on these bounds.
        polish=False,
    )

    print(f"Total generations used: {result.nit}")

    return (cartesian.vector_to_state(result.x, template), result)
    # return (spherical.vector_to_state(result.x, template), result)


def super_optimize_de(initial_state: State, seed: int, verbose=False):
    template = initial_state

    num_cams = len(template.cameras)

    cartesian = CartesianSerialize(seed)
    # spherical = SphericalSerialize(num_cams, num_faces, template.scale, seed)

    initial_vec = cartesian.state_to_vector(initial_state)
    # initial_vec = spherical.state_to_vector(initial_state)

    bounds = cartesian.init_bounds(template)
    # bounds = spherical.init_bounds()

    num_particles = 200
    init_pop = cartesian.init_pop(num_particles, initial_vec, bounds, num_cams)
    # init_pop = spherical.init_pop(
    #     num_particles, initial_vec, bounds, num_cams, num_faces
    # )

    def objective(vec):
        state = cartesian.vector_to_state(vec, template)
        # state = spherical.vector_to_state(vec, template)
        cost = total_cost(state, verbose)
        # render_from_state(None, state)
        return cost

    result = differential_evolution(
        objective,
        bounds,
        strategy="rand1bin",
        maxiter=5000,
        init=init_pop,
        mutation=(0.2, 0.7),
        popsize=num_particles,
        recombination=0.9,
        rng=seed,
        # Fitness Stagnation:
        # tol=0 ignores relative change in favor of atol
        tol=0,
        # atol matches epsilon (10^-6)
        atol=1e-2,
        # Ensure it checks for 50 consecutive generations
        # Note: SciPy's internal 'convergence' check varies slightly by version,
        # but setting polish=False ensures it stops strictly on these bounds.
        polish=False,
    )

    print(f"Total generations used: {result.nit}")

    return (cartesian.vector_to_state(result.x, template), result)
    # return (spherical.vector_to_state(result.x, template), result)
