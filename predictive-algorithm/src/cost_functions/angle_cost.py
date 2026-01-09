import math
from state import State
from utils import center_of_face, angle_from_point
from constant import BIG_M
from astropy.units import Quantity
import astropy.units as u


def horizontal_cost(hor_deg: Quantity[u.degree]) -> float:
    cost = 0

    hor_threshold_limit = 30.0
    hor_hard_limit = 45.0

    if hor_deg < hor_threshold_limit:
        cost = hor_deg
    elif hor_deg < hor_hard_limit:
        cost = hor_threshold_limit + (hor_deg - hor_threshold_limit) ** 2
    else:
        base_at_limit = (
            hor_threshold_limit + (hor_hard_limit - hor_threshold_limit) ** 2
        )
        cost = BIG_M + base_at_limit + (hor_deg - hor_hard_limit) * 5000

    return cost


def vertical_cost(ver_deg: Quantity[u.degree]) -> float:
    cost = 0

    ver_min_limit = 30.0
    ver_max_limit = 45.0

    if ver_deg < ver_min_limit:
        dist_below = ver_min_limit - ver_deg
        cost = BIG_M + (dist_below * 5000)
    elif ver_deg <= ver_max_limit:
        midpoint = (ver_min_limit + ver_max_limit) / 2  # 37.5
        cost = (ver_deg - midpoint) ** 2
    else:
        dist_above = ver_deg - ver_max_limit
        cost = BIG_M + (dist_above * 5000)

    return cost


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        face_center = center_of_face(cam_state.face)
        hor, ver = angle_from_point(face_center, cam_state.pos, cam_state.angle)
        hor_deg, ver_deg = abs(math.degrees(hor)), abs(math.degrees(ver))

        print(hor_deg, ver_deg)

        cost += horizontal_cost(hor_deg)
        cost += vertical_cost(ver_deg)

    return cost
