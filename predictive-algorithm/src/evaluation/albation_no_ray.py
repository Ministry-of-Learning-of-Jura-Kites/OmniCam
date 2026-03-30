import json
import math
from os import path
import time
from cost_functions import angle_cost, resolution_cost
import numpy as np
from state import CameraConfiguration, CameraState, State
import quaternion
from utils import (
    center_of_face,
)
import vtk
import pyvista as pv
from main import assign_faces
from algorithms import CartesianSerialize
from scipy.optimize import differential_evolution
from cost_functions.occlusion_cost import is_in_view
from env import env_settings
from dev.visualization import init_3d_scene, render_from_state

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


def brute_force_intersect(polydata, p1, p2, tolerance, points, cell_ids):
    """
    Fixed: Manually iterates every cell to find intersections closer than the target.
    """
    # closest_t = 1.0 means we only care about hits between p1 and p2.
    # We subtract a small epsilon to ignore the target surface itself.
    closest_t = 1.0 - (tolerance * 0.1)

    hit_point = [0.0, 0.0, 0.0]
    hit_cell_id = -1
    found_hit = False

    # Pre-allocate VTK mutable objects
    t = vtk.mutable(0.0)
    x = [0.0, 0.0, 0.0]
    pcoords = [0.0, 0.0, 0.0]
    sub_id = vtk.mutable(0)

    num_cells = polydata.GetNumberOfCells()
    for i in range(num_cells):
        cell = polydata.GetCell(i)

        # IntersectWithLine returns 1 if the line (p1, p2) hits the cell
        if cell.IntersectWithLine(p1, p2, tolerance, t, x, pcoords, sub_id):
            # Update only if this hit is closer than our current best
            # AND it's not the target point itself (t < 1.0)
            if t < closest_t:
                closest_t = float(t)
                hit_point = list(x)
                hit_cell_id = i
                found_hit = True

    if found_hit:
        points.InsertNextPoint(hit_point)
        cell_ids.InsertNextId(hit_cell_id)
        return 1

    return 0


def brute_force_occlusion(state: State, cam_state: CameraState, face):
    total_occlusion_cost = 0

    face_center = center_of_face(face)
    check_corners = [
        (corner * (100 - i * 10) / 100 + face_center * i * 10 / 100, i)
        for corner in face
        for i in range(0, 3)
    ]
    for corner, weight in check_corners:
        # 1. Frustum Check
        valid_coord, ndc_coord = is_in_view(corner, cam_state)
        if not valid_coord:
            dist_outside = np.max(np.abs(ndc_coord) - 1.0, initial=0)
            total_occlusion_cost += 500 * (weight + 1) + (dist_outside * 100)
            continue

        # 2. Brute Force Occlusion Check
        points = vtk.vtkPoints()
        cell_ids = vtk.vtkIdList()
        tolerance = 0.1

        # Calling our new manual function instead of the locator
        # Ensure state.full_scene_polydata contains the merged scene.gltf
        code = brute_force_intersect(
            state.gltf,
            cam_state.pos,
            corner,
            tolerance,
            points,
            cell_ids,
        )

        if code != 0:
            to_corner_dist = np.linalg.norm(corner - cam_state.pos)
            to_hit_dist = np.linalg.norm(np.array(points.GetPoint(0)) - cam_state.pos)

            if to_hit_dist < (to_corner_dist - tolerance):
                blocked_depth = to_corner_dist - to_hit_dist
                total_occlusion_cost += 1000 * (blocked_depth**2)

    return total_occlusion_cost


def brute_force_closest_point(polydata, query_point):
    """
    Manually iterates every cell to find the absolute closest point on the mesh.
    Returns (closest_coords, cell_id, dist_squared)
    """
    min_dist2 = float("inf")
    closest_coords = np.array([0.0, 0.0, 0.0])
    closest_cell_id = -1

    # Temporary variables for VTK math
    closest_p = [0.0, 0.0, 0.0]
    sub_id = vtk.reference(0)
    dist2 = vtk.reference(0.0)

    num_cells = polydata.GetNumberOfCells()
    for i in range(num_cells):
        cell = polydata.GetCell(i)
        # Find the closest point on THIS specific cell
        cell.EvaluatePosition(
            query_point, closest_p, sub_id, [0, 0, 0], dist2, [0, 0, 0]
        )

        current_dist2 = dist2.get()
        if current_dist2 < min_dist2:
            min_dist2 = current_dist2
            closest_coords = np.array(closest_p)
            closest_cell_id = i

    return closest_coords, closest_cell_id, min_dist2


def brute_force_mounting_cost(state: State, cam_state: CameraState, _face):
    # 1. Manual Brute Force Search
    # Instead of state.gltf_locator, we use our manual search function
    closest_p, cell_id, dist2 = brute_force_closest_point(state.gltf, cam_state.pos)

    d_val = math.sqrt(dist2)

    # Normalizing tolerances based on scene scale
    maximum_mounting_tolerance = 0.8 / state.scale
    minimum_mounting_tolerance = 0.3 / state.scale

    cost = 0.0

    # 2. Distance Penalty (Stay within the "sweet spot" of the surface)
    # Penalizes the camera if it's too far or too close to a wall/ceiling
    if d_val < minimum_mounting_tolerance or d_val > maximum_mounting_tolerance:
        # Using a quadratic pull to keep the optimization landscape smooth
        cost += 5000 * ((d_val - minimum_mounting_tolerance) ** 2)

    # 3. The "Below" Penalty (Hard Push away from floors)
    # If the closest surface point is below the camera center,
    # we treat it as a floor and penalize heavily.
    if closest_p[1] < cam_state.pos[1] - 0.001:
        # 20k flat penalty + gradient based on distance to ensure
        # the camera "climbs" out of the floor during optimization.
        cost += 20000 + (d_val * 1000)

    return cost


def total_cost_pair(state: State, cam_state: CameraState, face, verbose=False):
    angle = 0.18 * angle_cost.cost_single_cam(state, cam_state, face, verbose)
    resolution = 0.31 * resolution_cost.cost_single_cam(state, cam_state, face)
    occlusion = 0.37 * brute_force_occlusion(state, cam_state, face)
    mounting = 0.14 * brute_force_mounting_cost(state, cam_state, face)
    # occlusion = 0
    # mounting = 0
    return angle + resolution + occlusion + mounting, {
        "angle": angle,
        "res": resolution,
        "occ": occlusion,
        "mount": mounting,
    }


def total_cost(state: State, verbose: bool = False):
    # We want to know how well every camera sees every face
    # num_faces = len(state.faces)
    num_cams = len(state.cameras)
    if verbose:
        stats = {
            i: {"angle": [], "res": [], "occ": [], "mount": 0} for i in range(num_cams)
        }
    cost = 0

    for c_idx, cam in enumerate(state.cameras):
        for face in cam.faces:
            total, b = total_cost_pair(state, cam, face, verbose)
            if verbose:
                stats[c_idx]["angle"].append(b.get("angle", 0))
                stats[c_idx]["mount"] = b.get("mount", 0)
                stats[c_idx]["res"].append(b.get("res", 0))
                stats[c_idx]["occ"].append(b.get("occ", 0))
            cost += total
    if verbose:
        print(stats)

    return cost


def optimize_de(initial_state: State, seed: int, verbose=False):
    template = initial_state

    num_cams = len(template.cameras)

    # cartesian = CartesianSerialize()
    cartesian = CartesianSerialize(seed)

    initial_vec = cartesian.state_to_vector(initial_state)

    bounds = cartesian.init_bounds(template)

    num_particles = 50
    init_pop = cartesian.init_pop(num_particles, initial_vec, bounds, num_cams)

    def objective(vec):
        # state = cartesian.vector_to_state(vec, template)
        state = cartesian.vector_to_state(vec, template)
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
    return (cartesian.vector_to_state(result.x, template), result)


# --- Benchmarking Loop ---
seeds = range(2000, 2000 + 30)  # 2000 to 2014 inclusive
times = []
costs = []
results_gens = []

print(f"Starting benchmark for seeds {seeds.start} to {seeds.stop - 1}...")

for seed in seeds:

    pl = None
    if env_settings.dev_mode:
        from pyvistaqt import BackgroundPlotter

        pl = BackgroundPlotter()

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

    if env_settings.dev_mode:
        init_3d_scene(pl, state)
        render_from_state(pl, state)
        pl.show()
        breakpoint()

    print(f"Seed {seed} | Time: {elapsed_time:.4f}s | Cost: {current_cost:.4f}")

export = {
    "times": times,
    "costs": costs,
    "results_gens": results_gens,
    "seeds": list(seeds),
}


with open("albation no ray.json", "w") as json_file:
    # Step 5: Format the output with 4-space indentation
    json.dump(export, json_file, indent=4)
