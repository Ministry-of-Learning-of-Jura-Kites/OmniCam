import json
import math
from os import path
import time
from cost_functions import total_cost
from algorithms.differential_evolution import optimize_de, super_optimize_de
import numpy as np
from state import CameraConfiguration, CameraState, State
import quaternion
from utils import center_of_face
import vtk
import pyvista as pv
from main import assign_faces


# --- Static Setup ---
gltf_path = "/home/frook/Downloads/omnicam/test case a.glb"
gltf = pv.read(path.join(gltf_path)).combine().extract_surface().triangulate().clean()

gltf_locator = vtk.vtkStaticCellLocator()
gltf_locator.SetDataSet(gltf)
gltf_locator.BuildLocator()

face = np.array([[2, 2 + 10, 0], [-2, 2 + 10, 0], [-2, -2 + 10, 0], [2, -2 + 10, 0]])
cam_config = CameraConfiguration(
    pixels=[5000, 5000],
    vfov=60,
    name="f",
)

# --- Benchmarking Parameters ---
seeds = range(2000, 2000 + 30)
times = []
costs = []
results_gens = []

print(f"Benchmarking {len(seeds)} seeds...")

for seed in seeds:
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

    state = assign_faces(state, seed)

    start_time = time.perf_counter()

    # ASSUMPTION: optimize_de returns (final_state, generations)
    # If your function only returns final_state, you may need to
    # modify optimize_de to return the iteration count.
    final_state, result = super_optimize_de(state, seed)

    gens = result.nit

    end_time = time.perf_counter()

    elapsed = end_time - start_time
    cost = total_cost(final_state, True)

    times.append(elapsed)
    costs.append(cost)
    results_gens.append(gens)

    print(f"Seed {seed}: {gens} gens | {elapsed:.4f}s | Cost: {cost:.4f}")

export = {
    "times": times,
    "costs": costs,
    "results_gens": results_gens,
    "seeds": list(seeds),
}


with open("case a super.json", "w") as json_file:
    # Step 5: Format the output with 4-space indentation
    json.dump(export, json_file, indent=4)  #
