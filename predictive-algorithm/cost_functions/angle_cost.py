from state import State
from utils import center_of_face, angle_from_point


def cost(state: State):
    cost = 0
    for cam_state in state.cameras:
        face_center = center_of_face(cam_state.face)
        angle_from_point(face_center, cam_state.pos, cam_state.angle)
    return cost
