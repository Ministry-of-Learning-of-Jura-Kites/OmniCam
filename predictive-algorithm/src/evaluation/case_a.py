from os import path
import time
from cost_functions import total_cost
from algorithms.differential_evolution import optimize_de
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
results_time = []
results_cost = []
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
    final_state, result = optimize_de(state, seed)

    gens = result.nit

    end_time = time.perf_counter()

    elapsed = end_time - start_time
    cost = total_cost(final_state, True)

    results_time.append(elapsed)
    results_cost.append(cost)
    results_gens.append(gens)

    print(f"Seed {seed}: {gens} gens | {elapsed:.4f}s | Cost: {cost:.4f}")

# --- Final Statistics ---
avg_time = np.mean(results_time)
avg_cost = np.mean(results_cost)
med_cost = np.median(results_cost)

# Calculate a (mean) and b (standard deviation)
gen_mean = np.mean(results_gens)
gen_std = np.std(results_gens)

print("\n" + "=" * 40)
print(f"AVERAGE RESULTS (Seeds {seeds.start}-{seeds.stop-1})")
print(f"Generations:  {gen_mean:.2f} ± {gen_std:.2f}")
print(f"Avg Time:     {avg_time:.4f} seconds")
print(f"Avg Cost:     {avg_cost:.4f}")
print(f"Median Cost:     {med_cost:.4f}")
print("=" * 40)
