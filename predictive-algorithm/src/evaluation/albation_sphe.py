import json
from os import path
import time
from cost_functions import total_cost
import numpy as np
from state import CameraConfiguration, CameraState, State
import quaternion
from utils import (
    center_of_face,
)
import vtk
import pyvista as pv
from main import assign_faces
from algorithms import CartesianSerialize, SphericalSerialize
from scipy.optimize import differential_evolution

# --- Setup Shared Resources ---
gltf = (
    pv.read(path.join("/home/frook/Downloads/omnicam/test case b.glb"))
    .combine()
    .extract_surface()
    .triangulate()
    .clean()
)

gltf_locator = vtk.vtkStaticCellLocator()
gltf_locator.SetDataSet(gltf)
gltf_locator.BuildLocator()


face = np.array([[2, 2 + 1, 2], [-2, 2 + 1, 2], [-2, -2 + 1, 2], [2, -2 + 1, 2]])
cam_config = CameraConfiguration(
    pixels=[500, 500],
    vfov=60,
    name="f",
)


def optimize_de(initial_state: State, seed: int, verbose=False):
    template = initial_state

    num_faces = len(template.faces)
    num_cams = len(template.cameras)

    # cartesian = CartesianSerialize()
    spherical = SphericalSerialize(num_cams, num_faces, template.scale, seed)

    initial_vec = spherical.state_to_vector(initial_state)

    bounds = spherical.init_bounds()

    num_particles = 50
    init_pop = spherical.init_pop(num_particles, initial_vec, bounds, num_cams, None)

    def objective(vec):
        # state = cartesian.vector_to_state(vec, template)
        state = spherical.vector_to_state(vec, template)
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

    # return cartesian.vector_to_state(result.x, template)
    return (spherical.vector_to_state(result.x, template), result)


# --- Benchmarking Loop ---
seeds = range(2000, 2000 + 30)  # 2000 to 2014 inclusive
times = []
costs = []
results_gens = []

print(f"Starting benchmark for seeds {seeds.start} to {seeds.stop - 1}...")

for seed in seeds:
    # 1. Re-initialize the state for each seed to ensure a clean baseline
    state = State(
        faces=[face],
        face_to_cam=dict(),
        face_centers=[center_of_face(face)],
        cameras=[
            CameraState(
                faces=None,
                pos=[5, 0, 0],
                angle=quaternion.from_vector_part([0, 0, 0, 1]),
                center_of_faces=None,
                camera_config=cam_config,
                name="gg",
            )
        ],
        scale=1,
        gltf=gltf,
        gltf_locator=gltf_locator,
    )

    # 2. Assign faces based on the current seed
    state = assign_faces(state, seed)

    # 3. Run Optimization and Time it
    start_time = time.perf_counter()
    final_state, result = optimize_de(state, seed)
    end_time = time.perf_counter()

    # 4. Record Results
    elapsed_time = end_time - start_time
    current_cost = total_cost(final_state, True)

    times.append(elapsed_time)
    costs.append(current_cost)
    results_gens.append(result.nit)

    print(f"Seed {seed} | Time: {elapsed_time:.4f}s | Cost: {current_cost:.4f}")

export = {
    "times": times,
    "costs": costs,
    "results_gens": results_gens,
    "seeds": list(seeds),
}


with open("albation sphe.json", "w") as json_file:
    # Step 5: Format the output with 4-space indentation
    json.dump(export, json_file, indent=4)  #
