import base64
from os import path
import uuid
from env import env_settings
import vtk
import pyvista as pv
import numpy as np
import time


def parse_uuid_base64(base64_str: str) -> uuid.UUID:
    padding = "=" * (4 - len(base64_str) % 4)
    padded_str = base64_str + padding
    decoded_bytes = base64.urlsafe_b64decode(padded_str)
    return uuid.UUID(bytes=decoded_bytes)


def brute_force_intersect(mesh, p1, p2, tol=0.1):
    """Checks every cell in the mesh for an intersection manually."""
    for i in range(mesh.n_cells):
        cell = mesh.GetCell(i)
        t, x, pcoords, subId = (
            vtk.reference(0.0),
            [0.0, 0.0, 0.0],
            [0.0, 0.0, 0.0],
            vtk.reference(0),
        )
        # Using the cell's internal IntersectWithLine
        intersect_code = cell.IntersectWithLine(p1, p2, tol, t, x, pcoords, subId)
        if intersect_code == 1:
            return True
    return False


file_path_base = path.join(
    env_settings.model_file_path,
    "3d_models",
    "2f310ec5-bd86-4230-8051-4e60346d56ae",
)
# Model IDs
models = {
    "Small": "/home/frook/Downloads/living 1.14.glb",
    "Medium": "/home/frook/Downloads/living 5.57.glb",
    # "Large": "/home/frook/Downloads/living.glb",
    # "Super Large": "/home/frook/Downloads/supar-large-living.glb",
    # "Super Super Large": "/home/frook/Downloads/supar-super-large-living.glb",
}

num_tests = 50  # Reduced slightly for the 'Large' brute force safety
results = []

print(
    f"{'Size':<10} | {'Cells':<12} | {'Avg Locator (s)':<18} | {'Avg Brute (s)':<18} | {'Speedup'}"
)
print("-" * 85)

for label, file_path in models.items():
    # 1. Load and Pre-process
    mesh = pv.read(file_path).combine().extract_surface().triangulate().clean()
    n_cells = mesh.n_cells

    # 2. Build Locator
    gltf_locator = vtk.vtkStaticCellLocator()
    gltf_locator.SetDataSet(mesh)
    gltf_locator.BuildLocator()

    # 3. Benchmark Setup
    rng = np.random.default_rng(42)
    bounds = mesh.bounds

    def get_pt():
        return [
            rng.uniform(bounds[0], bounds[1]),
            rng.uniform(bounds[2], bounds[3]),
            rng.uniform(bounds[4], bounds[5]),
        ]

    loc_times = []
    brute_times = []

    for _ in range(num_tests):
        p1, p2 = get_pt(), get_pt()

        # Locator Test
        t0 = time.perf_counter()
        pts_loc, ids_loc = vtk.vtkPoints(), vtk.vtkIdList()
        gltf_locator.IntersectWithLine(p1, p2, 0.1, pts_loc, ids_loc)
        loc_times.append(time.perf_counter() - t0)

        # Brute Force Test
        t1 = time.perf_counter()
        brute_force_intersect(mesh, p1, p2)
        brute_times.append(time.perf_counter() - t1)

    avg_loc = np.mean(loc_times)
    avg_brute = np.mean(brute_times)
    speedup = avg_brute / avg_loc

    print(
        f"{label:<10} | {n_cells:<12,} | {avg_loc:<18.6f} | {avg_brute:<18.6f} | {speedup:.1f}x"
    )

    results.append(
        {
            "label": label,
            "cells": n_cells,
            "avg_loc": avg_loc,
            "avg_brute": avg_brute,
            "speedup": speedup,
        }
    )

print("\n--- Benchmark Complete ---")
