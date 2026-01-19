import functools
from state import CameraState, State
import numpy as np
from dataclasses import replace
from utils import center_of_face, look_at_quaternion
import quaternion


@functools.cache
def state_vector_dim():
    face = np.array(
        [
            [-1.0, -1.0, -1.0],
            [-1.0, -1.0, 1.0],
            [-1.0, 1.0, 1.0],
            [-1.0, 1.0, -1.0],
        ]
    )
    vec = state_to_vector(
        state=State(
            [
                CameraState(
                    face=face,
                    pos=np.array([0, 0, 0]),
                    angle=quaternion.from_float_array([0, 0, 0, 0]),
                    pixels=np.array([1920, 1080]),
                    vfov=70,
                )
            ],
            scale=1,
            gltf=None,
        )
    )

    return len(vec)


def state_to_vector(state: State):
    """Flattens State into a 1D numpy array."""
    vec = []
    for cam in state.cameras:
        vec.extend(cam.pos)  # Just 3 params: x, y, z
    return np.array(vec)


def vector_to_state(vec, template_state: State):
    """Reconstructs State from a 1D numpy array."""
    new_cameras = []
    idx = 0
    for i in range(len(template_state.cameras)):
        pos = vec[idx : idx + 3]

        # Calculate Look-At rotation automatically
        face_center = center_of_face(template_state.cameras[i].face)
        direction = face_center - pos
        # Use your existing utility to keep the camera pointed at the target
        angle = look_at_quaternion(direction)

        cam = replace(template_state.cameras[i], pos=pos, angle=angle)
        new_cameras.append(cam)
        idx += 3

    return State(
        cameras=new_cameras, scale=template_state.scale, gltf=template_state.gltf
    )


def state_to_spherical_vector(state: State):
    """
    Converts a State object back into a 1D array of [r, theta, phi] per camera.
    Inverse of spherical_vector_to_state.
    """
    vec = []
    for cam in state.cameras:
        face_center = center_of_face(cam.face)

        # 1. Get relative position (Camera - Face)
        rel_pos = cam.pos - face_center
        x, y, z = rel_pos

        # 2. Calculate Radius (r)
        r = np.linalg.norm(rel_pos)

        # 3. Calculate Elevation (phi)
        # phi is the angle from the XZ plane towards the Y axis
        # Using arcsin(y/r) matches your: local_y = r * sin(phi)
        phi_rad = np.arcsin(y / (r + 1e-8))
        phi_deg = np.degrees(phi_rad)

        # 4. Calculate Azimuth (theta)
        # Using atan2(z, x) matches your: local_x = r*cos(phi)*cos(theta)
        # and local_z = r*cos(phi)*sin(theta)
        theta_rad = np.arctan2(z, x)
        theta_deg = np.degrees(theta_rad)

        vec.extend([r, theta_deg, phi_deg])

    return np.array(vec)


def spherical_vector_to_state(vec, template: State):
    new_cameras = []
    for i in range(len(template.cameras)):
        r, theta, phi = vec[i * 3 : i * 3 + 3]

        # 1. Get face orientation
        face_center = center_of_face(template.cameras[i].face)
        # Assuming your face normal logic is available:
        # We calculate the position relative to the face center

        # Spherical to Cartesian (Standard math)
        theta_rad = np.radians(theta)
        phi_rad = np.radians(phi)

        local_x = r * np.cos(phi_rad) * np.cos(theta_rad)
        local_y = r * np.sin(phi_rad)
        local_z = r * np.cos(phi_rad) * np.sin(theta_rad)

        # 2. Offset from face center
        pos = face_center + np.array([local_x, local_y, local_z])

        # 3. Look-at logic (Automatically handles orientation)
        direction = face_center - pos
        angle = look_at_quaternion(direction)

        cam = replace(template.cameras[i], pos=pos, angle=angle)
        new_cameras.append(cam)

    return State(
        cameras=new_cameras,
        scale=template.scale,
        gltf=template.gltf,
        gltf_locator=template.gltf_locator,
    )
