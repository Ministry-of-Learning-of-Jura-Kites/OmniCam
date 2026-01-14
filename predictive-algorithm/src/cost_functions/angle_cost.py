import math

import numpy as np
import quaternion
from state import State
from utils import angle_from_face_normal, center_of_face
from constant import BIG_M
from astropy.units import Quantity
import astropy.units as u


def horizontal_cost(hor_deg: Quantity[u.degree]) -> float:
    cost = 0

    hor_deg = hor_deg.to_value(u.degree)
    hor_deg = abs(hor_deg)

    threshold = 30.0

    if hor_deg <= threshold:
        # Linear cost in the "good" zone
        return hor_deg
    else:
        # Quadratic penalty for going over.
        # Smooth at the join (30), but gets very steep very fast.
        # This gives the optimizer a clear 'gravity' back toward 30.
        return threshold + 50 * (hor_deg - threshold) ** 2

    return cost


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


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        hor, ver = angle_from_face_normal(
            cam_state.face, cam_state.pos, cam_state.angle
        )
        hor_deg, ver_deg = hor.to(u.degree), ver.to(u.degree)

        # print(hor_deg, ver_deg)

        cost += horizontal_cost(hor_deg)
        cost += vertical_cost(ver_deg)

    # print("angle cost:", cost)

    return cost
