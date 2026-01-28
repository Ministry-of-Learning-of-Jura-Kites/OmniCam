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
        # TODO: Support face choosing
        face_center = center_of_face(state.faces[0])

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

        # 1. Pivot point: Instead of a specific face, use the scene center
        # or the initial position provided in the template.
        pivot = np.array([0, 0, 0])  # Scene origin, or use a specific landmark

        # 2. Convert Spherical to World Cartesian
        theta_rad = np.radians(theta)
        phi_rad = np.radians(phi)

        x = r * np.cos(phi_rad) * np.cos(theta_rad)
        y = r * np.sin(phi_rad)
        z = r * np.cos(phi_rad) * np.sin(theta_rad)
        pos = pivot + np.array([x, y, z])

        # 3. Dynamic Look-At:
        # Since we don't know which face it's looking at yet,
        # we look at the 'Closest Face' or the 'Average' of face cluster
        # Or better: let the cost function determine the best angle.

        # For now, let's look at the nearest face center to establish an orientation
        face_centers = np.array([center_of_face(f) for f in template.faces])
        dists = np.linalg.norm(face_centers - pos, axis=1)
        target_face_center = face_centers[np.argmin(dists)]

        direction = target_face_center - pos
        angle = look_at_quaternion(direction)

        # Update the camera. Note: we might store 'None' in cam.face
        # because the face-to-camera link is now calculated in total_cost()
        cam = replace(template.cameras[i], pos=pos, angle=angle)
        new_cameras.append(cam)

    return replace(template, cameras=new_cameras)
