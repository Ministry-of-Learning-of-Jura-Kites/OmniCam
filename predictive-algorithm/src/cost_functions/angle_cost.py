import math

import numpy as np
import quaternion
from state import CameraState, State
from utils import angle_from_face_normal, center_of_face
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


def cost_single_cam(state: State, cam_state: CameraState, face: Array4x3):
    cost = 0
    hor, ver = angle_from_face_normal(face, cam_state.pos, cam_state.angle)
    hor_deg, ver_deg = hor.to(u.degree), ver.to(u.degree)

    print(hor_deg, ver_deg)

    cost += horizontal_cost(hor_deg)
    cost += vertical_cost(ver_deg)

    return cost


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        cost += cost_single_cam(cam_state, cam_state.face)
    return cost
