import json
import math
from os import path
import time
from cost_functions import total_cost
from algorithms.uniform_random import optimize_random_search
import numpy as np
from state import CameraConfiguration, CameraState, State
import quaternion
from utils import (
    center_of_face,
)
from env import env_settings
import vtk
from dev.visualization import init_3d_scene, render_from_state
import pyvista as pv
from main import assign_faces

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

# --- Benchmarking Loop ---
seeds = range(2000, 2000 + 10)  # 2000 to 2014 inclusive
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
    final_state, result = optimize_random_search(state, seed)
    end_time = time.perf_counter()

    # 4. Record Results
    elapsed_time = end_time - start_time
    current_cost = total_cost(final_state, True)

    times.append(elapsed_time)
    costs.append(current_cost)
    results_gens.append(result)

    print(f"Seed {seed} | Time: {elapsed_time:.4f}s | Cost: {current_cost:.4f}")

export = {
    "times": times,
    "costs": costs,
    "results_gens": results_gens,
    "seeds": list(seeds),
}

with open("case b rand.json", "w") as json_file:
    # Step 5: Format the output with 4-space indentation
    json.dump(export, json_file, indent=4)  #
