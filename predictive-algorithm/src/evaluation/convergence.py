import json
import time
import numpy as np
from main import OptimizeRequest, assign_faces
import uuid
import base64
from env import env_settings
from state import State
import pyvista as pv
from os import path
import vtk
from main import transform_faces, transform_cameras
from utils import center_of_face
from cost_functions import total_cost
from scipy.optimize import differential_evolution
from algorithms import SphericalSerialize
from state import State
from cost_functions import total_cost

import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
import pandas as pd


def super_optimize_de(
    initial_state: State, seed: int, maxiter: int, num_particles: int, verbose=False
):
    template = initial_state

    convergence_history = []

    num_faces = len(template.faces)
    num_cams = len(template.cameras)

    # cartesian = CartesianSerialize()
    spherical = SphericalSerialize(num_cams, num_faces, template.scale, seed)

    # initial_vec = cartesian.state_to_vector(initial_state)
    initial_vec = spherical.state_to_vector(initial_state)

    # bounds = cartesian.init_bounds(template)
    bounds = spherical.init_bounds()

    # init_pop = cartesian.init_pop(num_particles, initial_vec, bounds, num_cams)
    init_pop = spherical.init_pop(
        num_particles, initial_vec, bounds, num_cams, num_faces
    )

    def objective(vec):
        # state = cartesian.vector_to_state(vec, template)
        state = spherical.vector_to_state(vec, template)
        cost = total_cost(state, verbose)
        # render_from_state(None, state)
        return cost

    def callback(xk, convergence):
        # xk is the best parameter vector found so far in this generation
        # We re-calculate the objective to get the exact cost for the graph
        current_best_cost = objective(xk)
        convergence_history.append(1 / current_best_cost)

    result = differential_evolution(
        objective,
        bounds,
        strategy="rand1bin",
        maxiter=maxiter,
        init=init_pop,
        mutation=(0.2, 0.7),
        popsize=num_particles,
        recombination=0.9,
        rng=seed,
        # Fitness Stagnation:
        # tol=0 ignores relative change in favor of atol
        tol=0,
        # atol matches epsilon (10^-6)
        atol=1e-4,
        callback=callback,
        # Ensure it checks for 50 consecutive generations
        # Note: SciPy's internal 'convergence' check varies slightly by version,
        # but setting polish=False ensures it stops strictly on these bounds.
        polish=False,
    )

    print(f"Total generations used: {result.nit}")

    # return cartesian.vector_to_state(result.x, template)
    return (spherical.vector_to_state(result.x, template), result, convergence_history)


def optimize(
    req: OptimizeRequest, seed: int, maxiter: int, num_particles: int
) -> State:
    pl = None
    if env_settings.dev_mode:
        from pyvistaqt import BackgroundPlotter

        pl = BackgroundPlotter()

    gltf = (
        pv.read(
            path.join(
                env_settings.model_file_path,
                "3d_models",
                req.project_id,
                req.model_id + ".glb",
            )
        )
        .combine()
        .extract_surface()
        .triangulate()
        .clean()
    )

    gltf_locator = vtk.vtkStaticCellLocator()
    gltf_locator.SetDataSet(gltf)
    gltf_locator.BuildLocator()

    faces = transform_faces(req.faces)
    cameras = transform_cameras(req.cam_configs)
    state = State(
        faces=faces,
        face_to_cam=dict(),
        face_centers=list(map(center_of_face, faces)),
        cameras=cameras,
        scale=req.scale,
        gltf=gltf,
        gltf_locator=gltf_locator,
    )

    num_faces = len(state.faces)
    num_cameras = len(state.cameras)

    if num_cameras > num_faces:
        return

    state = assign_faces(state, seed)

    start_time = time.perf_counter()
    # from cost_functions import total_cost
    # print(total_cost(state))
    # breakpoint()

    # final_state = optimize_pso(
    #     state,
    #     # pl,
    #     None,
    # )
    final_state, _res, convergence_history = super_optimize_de(
        state, seed, maxiter, num_particles
    )

    end_time = time.perf_counter()
    elapsed_time = end_time - start_time
    print(f"Elapsed time: {elapsed_time:.4f} seconds")

    total = total_cost(final_state, True)
    print("total cost: ", total)

    if env_settings.dev_mode:
        pl.close()

    return total, convergence_history


def parse_uuid_base64(base64_str: str) -> uuid.UUID:
    """
    Direct Python port of the Go ParseUuidBase64 function.
    Takes a 22-character Base64 string and returns a UUID object.
    """
    # 1. Add back the missing '=' padding
    # Base64 strings must have a length divisible by 4
    padding = "=" * (4 - len(base64_str) % 4)
    padded_str = base64_str + padding

    # 2. Decode using the URL-safe alphabet (+ -> -, / -> _)
    decoded_bytes = base64.urlsafe_b64decode(padded_str)

    # 3. Convert the 16 raw bytes into a Python UUID object
    return uuid.UUID(bytes=decoded_bytes)


small_model_id_raw = "aBOpEuNkQZCQQeOnct1oRA"
medium_model_id_raw = "9pUzFr5iRWeIcu50urUsvQ"
large_model_id_raw = "It6QmaHGSHCaTtiGSS13mA"
small_model_id = parse_uuid_base64(small_model_id_raw)
medium_model_id = parse_uuid_base64(medium_model_id_raw)
large_model_id = parse_uuid_base64(large_model_id_raw)

trials = 10
results_summary = []
op_fitness_histories = []
ref_fitness_histories = []


# Helper to flip Cost to Fitness
def cost_to_fitness(cost_value):
    return 1 / cost_value


for model_id in [small_model_id]:
    op_fitness_scores = []
    ref_fitness_scores = []
    op_times = []
    ref_times = []

    for i in range(trials):
        seed = i + 2000
        with open("src/evaluation/ff.json", "r") as file:
            req = json.loads(file.read())
            data = json.loads(req["data"])
            data["model_id"] = str(model_id)
        payload = OptimizeRequest.model_validate(data)

        # 1. Run Operational Trial
        start_op = time.perf_counter()
        cost_op, h_op = optimize(payload, seed, 500, 50)
        op_fitness_histories.append(h_op)
        op_fitness_scores.append(cost_to_fitness(cost_op))
        op_times.append(time.perf_counter() - start_op)

        # 2. Run Reference Trial
        start_ref = time.perf_counter()
        cost_ref, h_ref = optimize(payload, seed, 5000, 200)
        ref_fitness_histories.append(h_ref)
        ref_fitness_scores.append(cost_to_fitness(cost_ref))
        ref_times.append(time.perf_counter() - start_ref)

    # Calculate Metrics using Fitness (Higher is Better)
    f_ref_max = np.max(ref_fitness_scores)
    f_op_avg = np.mean(op_fitness_scores)
    eta = (f_op_avg / f_ref_max) * 100
    avg_op_t = np.mean(op_times)
    avg_ref_t = np.mean(ref_times)

    results_summary.append(
        {
            "model": str(model_id),
            "f_ref_max": f_ref_max,
            "f_op_avg": f_op_avg,
            "eta_percent": eta,
            "avg_op_time": avg_op_t,
            "avg_ref_time": avg_ref_t,
            "speedup_factor": np.mean(ref_times) / np.mean(op_times),
        }
    )


# --- Fixed JSON Export ---
export_data = {
    "summary": results_summary,
    "metadata": {
        "op_config": {"G": 500, "Np": 50},
        "ref_config": {"G": 5000, "Np": 200},
        "trials": trials,
        "objective": "Fitness (Inversed Cost)",
    },
    # Store histories as raw costs (preserving original data)
    "op_fitness_histories": [list(h) for h in op_fitness_histories],
    "ref_fitness_histories": [list(h) for h in ref_fitness_histories],
}

breakpoint()

with open("optimization_results_v1.json", "w") as f:
    json.dump(export_data, f, indent=4)
