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

    hor_threshold_limit = 30.0
    hor_hard_limit = 45.0

    if hor_deg < hor_threshold_limit:
        cost = hor_deg
    elif hor_deg < hor_hard_limit:
        cost = hor_threshold_limit + (hor_deg - hor_threshold_limit) * 1000
    else:
        base_at_limit = (
            hor_threshold_limit + (hor_hard_limit - hor_threshold_limit) * 1000
        )
        cost = BIG_M + base_at_limit + (hor_deg - hor_hard_limit) * 5000

    return cost


def vertical_cost(ver_deg: Quantity[u.degree]) -> float:
    val = ver_deg.to_value(u.deg)

    ver_min_limit = 30.0
    ver_max_limit = 45.0
    midpoint = (ver_min_limit + ver_max_limit) / 2  # 37.5

    # Calculate the cost at the boundary to ensure continuity
    # (30 - 37.5)^2 = 56.25
    boundary_cost = (ver_min_limit - midpoint) ** 2

    if val < ver_min_limit:
        # Distance below the minimum
        dist_below = ver_min_limit - val
        return boundary_cost + BIG_M + (dist_below * 5000)

    elif val <= ver_max_limit:
        # Smooth parabolic cost for the "safe" zone
        return (val - midpoint) ** 2

    else:
        # Distance above the maximum
        dist_above = val - ver_max_limit
        return boundary_cost + BIG_M + (dist_above * 5000)


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        hor, ver = angle_from_face_normal(
            cam_state.face, cam_state.pos, cam_state.angle
        )
        hor_deg, ver_deg = hor.to(u.degree), ver.to(u.degree)

        print(hor_deg, ver_deg)

        cost += horizontal_cost(hor_deg)
        cost += vertical_cost(ver_deg)

    print("angle cost:", cost)

    return cost
