import math

import numpy as np
import quaternion
from state import CameraState, State
from utils import angle_from_face_normal, angle_from_face_position, center_of_face
from constant import BIG_M
from astropy.units import Quantity
import astropy.units as u
import logging
from basic_types import Array4x3


def horizontal_cost(hor_deg: Quantity[u.degree]) -> float:
    hor_deg = hor_deg.to_value(u.degree)
    hor_deg = abs(hor_deg)

    threshold = 30.0

    if hor_deg <= threshold:
        # Linear cost in the "good" zone
        return hor_deg * 50
    else:
        # Quadratic penalty for going over.
        # Smooth at the join (30), but gets very steep very fast.
        # This gives the optimizer a clear 'gravity' back toward 30.
        return threshold + 50 * (hor_deg - threshold) ** 2


def vertical_cost(ver_deg: Quantity[u.degree]) -> float:
    val = ver_deg.to_value(u.deg)

    ver_min = 30.0
    ver_max = 45.0
    midpoint = (ver_min + ver_max) / 2  # 37.5

    if ver_min <= val <= ver_max:
        return (val - midpoint) ** 2

    # Outside the limits:
    # We take the cost at the boundary and add a steep penalty.
    # This stays continuous but becomes much more expensive.
    if val < ver_min:
        return (ver_min - midpoint) ** 2 + 100 * (ver_min - val) ** 2
    else:
        return (ver_max - midpoint) ** 2 + 100 * (val - ver_max) ** 2


def off_center_penalty(
    h_off: Quantity[u.deg], v_off: Quantity[u.deg], h_fov=90.0, v_fov=60.0
) -> float:
    h = abs(h_off.to_value(u.deg))
    v = abs(v_off.to_value(u.deg))

    # 1. Internal Cost: Faces inside the frame should be centered
    # We use a gentle quadratic here.
    cost = (h**2 + v**2) * 5

    # 2. External 'Gravity' Cost:
    # If the face is outside the FOV (h > h_fov/2), we add a massive
    # but sloped penalty.
    h_limit = h_fov / 2
    v_limit = v_fov / 2

    if h > h_limit:
        # The further away it is, the harder it 'pulls' the camera back
        cost += 1000 * (h - h_limit) ** 2

    if v > v_limit:
        cost += 1000 * (v - v_limit) ** 2

    return cost


def cost_single_cam(
    state: State, cam_state: CameraState, face: Array4x3, verbose=False
):
    cost = 0
    hor, ver = angle_from_face_normal(face, cam_state.pos, cam_state.angle)
    hor_deg, ver_deg = hor.to(u.degree), ver.to(u.degree)
    cost += horizontal_cost(hor_deg)
    cost += vertical_cost(ver_deg)

    off_h, off_v = angle_from_face_position(face, cam_state.pos, cam_state.angle)
    off_h_deg, off_v_deg = off_h.to(u.deg), off_v.to(u.deg)
    cost += off_center_penalty(
        off_h_deg,
        off_v_deg,
        cam_state.camera_config.get_hfov(),
        cam_state.camera_config.vfov,
    )

    if verbose:
        print("hor_deg:", hor_deg)
        print("ver_deg:", ver_deg)
        print("off_h_deg:", off_h_deg)
        print("off_v_deg:", off_v_deg)

    # 3. FIELD OF VIEW (Hard Cutoff)
    # If the face is outside the camera's FOV (e.g. > 45 degrees off-center for a 90 FOV)
    # add a BIG_M penalty immediately.
    if abs(off_h.to_value(u.deg)) > 45 or abs(off_v.to_value(u.deg)) > 30:
        cost += 100000

    return cost


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        cost += cost_single_cam(cam_state, cam_state.face)
    return cost
