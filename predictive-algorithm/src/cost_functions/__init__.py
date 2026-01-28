import logging
from typing import List

import numpy as np
from state import CameraState, State
from . import angle_cost, resolution_cost, occlusion_cost, mounting_cost
from basic_types import Array3, Array4x3
from constant import BIG_M
from utils import center_of_face


def prune_faces_by_distance(
    cam_state: CameraState,
    all_faces: List[Array4x3],
    face_centers: List[Array3],
    max_dist=20.0,
):
    """
    Returns indices of faces that are within range and generally
    in front of the camera.
    """
    if not all_faces:
        return []

    # Calculates distance from camera to every face at once
    dists = np.linalg.norm(face_centers - cam_state.pos, axis=1)

    # Even before a full 'is_in_view', we can check the dot product
    # to see if the face is behind the camera's sensor plane.
    cam_forward = cam_state.forward_vector()  # e.g., result of your look_at_quaternion
    vecs_to_faces = face_centers - cam_state.pos

    # Normalize vectors to faces
    norms = np.linalg.norm(vecs_to_faces, axis=1, keepdims=True)
    unit_vecs_to_faces = vecs_to_faces / (norms + 1e-8)

    # Dot product > 0 means the face is 'in front' of the camera
    dots = np.dot(unit_vecs_to_faces, cam_forward)

    # Keep faces that are within max_dist AND at least slightly in front
    mask = (dists < max_dist) & (dots > -0.2)  # -0.2 allows for wide-angle peripheral

    return np.where(mask)[0]


def total_cost(state: State, verbose: bool = False):
    # We want to know how well every camera sees every face
    num_faces = len(state.faces)
    num_cams = len(state.cameras)
    score_matrix = np.full((num_faces, num_cams), BIG_M)

    if verbose:
        stats = {
            i: {"angle": [], "res": [], "occ": [], "mount": 0} for i in range(num_cams)
        }

    for c_idx, cam in enumerate(state.cameras):
        # OPTIMIZATION: Only check faces near this camera
        potential_faces = prune_faces_by_distance(cam, state.faces, state.face_centers)

        for f_idx in potential_faces:
            total, _ = total_cost_pair(state, cam, state.faces[f_idx])
            score_matrix[f_idx][c_idx] = total

    # For each face, we only care about the MINIMUM cost across all cameras.
    # This naturally allows one camera to be the 'best' for 10 faces at once.
    best_costs_per_face = np.min(score_matrix, axis=1)
    best_cam_indices = np.argmin(score_matrix, axis=1)

    # # 4. ACTIVE CAMERA PENALTY
    # # Identify which cameras are actually 'winners' for at least one face
    # winning_camera_indices = np.argmin(score_matrix, axis=1)
    # # Filter out faces that are totally occluded (cost == BIG_M)
    # active_indices = winning_camera_indices[best_costs_per_face < BIG_M]
    # num_active_cams = len(np.unique(active_indices))

    if verbose:
        for f_idx, c_idx in enumerate(best_cam_indices):
            if best_costs_per_face[f_idx] < BIG_M:
                # Re-run or retrieve the specific breakdown for the winner
                _, b = total_cost_pair(state, state.cameras[c_idx], state.faces[f_idx])
                stats[c_idx]["angle"].append(b.get("angle", 0))
                stats[c_idx]["res"].append(b.get("res", 0))
                stats[c_idx]["occ"].append(b.get("occ", 0))

        log_detailed_distribution(stats)

    return np.sum(best_costs_per_face)


def total_cost_pair(state: State, cam_state: CameraState, face: Array4x3):
    angle = angle_cost.cost_single_cam(state, cam_state, face)
    resolution = resolution_cost.cost_single_cam(state, cam_state, face)
    occlusion = occlusion_cost.cost_single_cam(state, cam_state, face)
    mounting = 0.5 * mounting_cost.cost_single_cam(state, cam_state, face)
    return angle + resolution + occlusion + mounting, {
        "angle": angle,
        "res": resolution,
        "occ": occlusion,
        "mount": mounting,
    }


def log_detailed_distribution(stats):
    print(
        f"\n{'Cam':<5} | {'Faces':<5} | {'Mounting':<10} | {'Avg Angle':<10} | {'Avg Res':<10} | {'Occl.':<8}"
    )
    print("-" * 70)
    for c_idx, data in stats.items():
        num_f = len(data["angle"])
        avg_ang = np.mean(data["angle"]) if num_f > 0 else 0
        avg_res = np.mean(data["res"]) if num_f > 0 else 0
        avg_occ = np.mean(data["occ"]) if num_f > 0 else 0

        print(
            f"{c_idx:<5} | {num_f:<5} | {data['mount']:<10.1f} | {avg_ang:<10.1f} | {avg_res:<10.1f} | {avg_occ:<8.1f}"
        )
