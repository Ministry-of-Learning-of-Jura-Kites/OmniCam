from abc import ABC, abstractmethod
import functools
from typing import List, Tuple
from state import CameraState, State
import numpy as np
from dataclasses import dataclass, replace
from utils import center_of_face, center_of_faces, look_at_quaternion
import quaternion
from cost_functions import total_cost
from constant import VectorSerialization, VECTOR_SERIALIZATION


class AlgorithmSerialization(ABC):
    @abstractmethod
    def init_pop(
        self,
    ) -> np.array:
        pass

    @abstractmethod
    def init_bounds(
        self,
    ) -> List[Tuple[float, float]]:
        pass

    @abstractmethod
    def init_bounds(
        self,
    ) -> List[Tuple[float, float]]:
        pass

    @abstractmethod
    def state_to_vector(self, state: State) -> np.array:
        pass

    @abstractmethod
    def vector_to_state(self, vec: np.array, template: State) -> State:
        pass


# def generic_state_to_vector(state: State):
#     match VECTOR_SERIALIZATION:
#         case VectorSerialization.CARTESIAN:
#             return state_to_cartesian_vector(state)
#         case VectorSerialization.SPHERICAL:
#             return state_to_spherical_vector(state)


# def generic_vector_to_state(vec, template: State):
#     match VECTOR_SERIALIZATION:
#         case VectorSerialization.CARTESIAN:
#             return cartesian_vector_to_state(vec, template)
#         case VectorSerialization.SPHERICAL:
#             return spherical_vector_to_state(vec, template)


# @functools.cache
# def state_spherical_vector_dim():
#     face = np.array(
#         [
#             [-1.0, -1.0, -1.0],
#             [-1.0, -1.0, 1.0],
#             [-1.0, 1.0, 1.0],
#             [-1.0, 1.0, -1.0],
#         ]
#     )
#     vec = state_to_spherical_vector(
#         state=State(
#             [
#                 CameraState(
#                     face=face,
#                     pos=np.array([0, 0, 0]),
#                     angle=quaternion.from_float_array([0, 0, 0, 0]),
#                     pixels=np.array([1920, 1080]),
#                     vfov=70,
#                 )
#             ],
#             scale=1,
#             gltf=None,
#         )
#     )

#     return len(vec)


# @functools.cache
# def state_cartesian_vector_dim():
#     face = np.array(
#         [
#             [-1.0, -1.0, -1.0],
#             [-1.0, -1.0, 1.0],
#             [-1.0, 1.0, 1.0],
#             [-1.0, 1.0, -1.0],
#         ]
#     )
#     vec = state_to_cartesian_vector(
#         state=State(
#             [
#                 CameraState(
#                     face=face,
#                     pos=np.array([0, 0, 0]),
#                     angle=quaternion.from_float_array([0, 0, 0, 0]),
#                     pixels=np.array([1920, 1080]),
#                     vfov=70,
#                 )
#             ],
#             scale=1,
#             gltf=None,
#         )
#     )

#     return len(vec)

# def state_to_spherical_vector(state: State):
#     """
#     Converts a State object back into a 1D array of [r, theta, phi] per camera.
#     Inverse of spherical_vector_to_state.
#     """
#     vec = []
#     for cam in state.cameras:
#         face_center = center_of_faces(cam.faces)

#         # 1. Get relative position (Camera - Face)
#         rel_pos = cam.pos - face_center
#         x, y, z = rel_pos

#         # 2. Calculate Radius (r)
#         r = np.linalg.norm(rel_pos)

#         # 3. Calculate Elevation (phi)
#         # phi is the angle from the XZ plane towards the Y axis
#         # Using arcsin(y/r) matches your: local_y = r * sin(phi)
#         phi_rad = np.arcsin(y / (r + 1e-8))
#         phi_deg = np.degrees(phi_rad)

#         # 4. Calculate Azimuth (theta)
#         # Using atan2(z, x) matches your: local_x = r*cos(phi)*cos(theta)
#         # and local_z = r*cos(phi)*sin(theta)
#         theta_rad = np.arctan2(z, x)
#         theta_deg = np.degrees(theta_rad)

#         vec.extend([r, theta_deg, phi_deg])

#     return np.array(vec)


# def spherical_vector_to_state(vec, template: State):
#     new_cameras = []
#     for i in range(len(template.cameras)):
#         r, theta, phi = vec[i * 3 : i * 3 + 3]

#         # 1. Pivot point: Instead of a specific face, use the scene center
#         # or the initial position provided in the template.
#         pivot = np.array([22.0, 1.6, 2])  # Scene origin, or use a specific landmark

#         # 2. Convert Spherical to World Cartesian
#         theta_rad = np.radians(theta)
#         phi_rad = np.radians(phi)

#         x = r * np.cos(phi_rad) * np.cos(theta_rad)
#         y = r * np.sin(phi_rad)
#         z = r * np.cos(phi_rad) * np.sin(theta_rad)
#         pos = pivot + np.array([x, y, z])

#         # 3. Dynamic Look-At:
#         # Since we don't know which face it's looking at yet,
#         # we look at the 'Closest Face' or the 'Average' of face cluster
#         # Or better: let the cost function determine the best angle.

#         # For now, let's look at the nearest face center to establish an orientation
#         face_centers = np.array([center_of_face(f) for f in template.faces])
#         dists = np.linalg.norm(face_centers - pos, axis=1)
#         target_face_center = face_centers[np.argmin(dists)]

#         direction = target_face_center - pos
#         angle = look_at_quaternion(direction)

#         # Delete face selecting information for next time to prevent confusion
#         cam = replace(template.cameras[i], pos=pos, angle=angle, faces=None)
#         new_cameras.append(cam)

#     return replace(template, cameras=new_cameras)


class CartesianSerialize(AlgorithmSerialization):
    def state_to_vector(self, state: State):
        """Flattens State into a 1D numpy array."""
        vec = []
        for cam in state.cameras:
            vec.extend(cam.pos)  # Just 3 params: x, y, z
        return np.array(vec)

    def vector_to_state(self, vec, template_state: State):
        """Reconstructs State from a 1D numpy array."""
        new_cameras = []
        idx = 0
        for i in range(len(template_state.cameras)):
            pos = vec[idx : idx + 3]

            # Calculate Look-At rotation automatically
            face_center = center_of_face(template_state.faces[0])
            direction = face_center - pos
            # Use your existing utility to keep the camera pointed at the target
            angle = look_at_quaternion(direction)

            cam = replace(template_state.cameras[i], pos=pos, angle=angle)
            new_cameras.append(cam)
            idx += 3

        return replace(template_state, cameras=new_cameras)

    def init_bounds(self, template: State) -> List[Tuple[float, float]]:
        # Define bounds
        bounds = []
        for _ in range(len(template.cameras)):
            bounds.extend([(-100 / template.scale, 100 / template.scale)] * 3)  # Pos
            # bounds.extend([(-1, 1)] * 4)  # Quaternion components

        return bounds

    def init_pop(self, num_particles, initial_vec, bounds, num_cams):
        """Initializes population for [x, y, z] vectorization."""
        total_dim = len(initial_vec)
        init_pop = np.empty((num_particles, total_dim))

        low_b = np.array([b[0] for b in bounds])
        high_b = np.array([b[1] for b in bounds])

        for i in range(num_particles):
            # Cartesian diversification is purely noise-based or random offsets
            particle = initial_vec.copy()

            # Add spatial noise
            noise = np.random.uniform(-5, 5, total_dim)

            # For the second half, we can add a larger "jump" to spread them out
            if i >= num_particles // 2:
                noise *= 3

            init_pop[i] = np.clip(particle + noise, low_b, high_b)

        return init_pop


@dataclass
class SphericalSerialize(AlgorithmSerialization):
    num_cams: int
    num_faces: int
    scale: float  # virtual meter per meter

    def init_bounds(self) -> List[Tuple[float, float]]:
        bounds = []
        for _ in range(self.num_cams):
            bounds.extend(
                [
                    (1.0 / self.scale, 25.0 / self.scale),  # Radius
                    (-180.0, 180.0),  # Theta (Azimuth)
                    (-45.0, 45.0),  # Phi (Elevation)
                ]
            )
        return bounds

    def init_pop(self, num_particles, initial_vec, bounds, num_cams, num_faces):
        """Initializes population with Normal-vector bias and radius diversification."""
        total_dim = len(initial_vec)
        vector_size = 3
        init_pop = np.empty((num_particles, total_dim))

        low_b = np.array([b[0] for b in bounds])
        high_b = np.array([b[1] for b in bounds])

        # Extract target "normal" or reference angles from the initial_vec
        # We assume initial_vec represents a 'good' starting orientation
        for i in range(num_particles):
            particle = initial_vec.copy()

            for cam_idx in range(num_cams):
                base = cam_idx * vector_size

                # 1. Diversify Radius (r)
                # Mix of close and far within bounds
                r_min, r_max = bounds[base][0], bounds[base][1]
                # Use a beta distribution to favor the middle-range but allow extremes
                particle[base] = np.random.uniform(r_min, r_max)

                # 2. Diversify Angles (theta, phi) with Normal Bias
                # Instead of pure uniform, we use a normal dist centered on the initial_vec
                # but with increasing spread as 'i' increases
                spread_factor = (
                    i / num_particles
                ) * 2.0  # Increases variance across pop

                # Theta (Azimuth) noise
                theta_noise = np.random.normal(0, 30 * spread_factor)
                particle[base + 1] += theta_noise

                # Phi (Elevation) noise - favor 'top-down' or 'front-on' based on initial
                phi_noise = np.random.normal(0, 15 * spread_factor)
                particle[base + 2] += phi_noise

                # Wrap angles
                particle[base + 1] = (particle[base + 1] + 180) % 360 - 180
                particle[base + 2] = np.clip(particle[base + 2], -90, 90)

            # 3. Add a final layer of uniform jitter for global coverage
            jitter = np.random.uniform(-2, 2, total_dim)

            # Ensure the particle stays within search space
            init_pop[i] = np.clip(particle + jitter, low_b, high_b)

        return init_pop

    def vector_to_state(self, vec, template: State):
        new_cameras = []
        vector_size = 3

        for i in range(len(template.cameras)):
            # We now use 4 parameters per camera
            r, theta, phi = vec[i * vector_size : i * vector_size + vector_size]

            # 1. Selection: Map the continuous DE variable to a valid face index
            # We use clip and round to handle the DE's floating point nature
            pivot = template.cameras[i].center_of_faces

            # 2. Local Spherical to World Cartesian
            theta_rad = np.radians(theta)
            phi_rad = np.radians(phi)

            local_x = r * np.cos(phi_rad) * np.cos(theta_rad)
            local_y = r * np.sin(phi_rad)
            local_z = r * np.cos(phi_rad) * np.sin(theta_rad)

            pos = pivot + np.array([local_x, local_y, local_z])

            # 3. Look-At Logic
            # It's safest to look back at the pivot (the face) we are orbiting
            direction = pivot - pos
            angle = look_at_quaternion(direction)

            cam = replace(template.cameras[i], pos=pos, angle=angle)
            new_cameras.append(cam)

        return replace(template, cameras=new_cameras)

    def state_to_vector(self, state: State):
        """
        Converts State back to [r, theta, phi, face_idx] per camera.
        Uses the nearest face as the anchor point.
        """
        vec = []

        for cam in state.cameras:
            # 1. Find the best anchor (nearest face)
            # dists = np.linalg.norm(face_centers - cam.pos, axis=1)
            # face_idx = np.argmin(dists)
            pivot = cam.center_of_faces

            # 2. Get relative position
            rel_pos = cam.pos - pivot
            x, y, z = rel_pos

            # 3. Calculate Radius (r)
            r = np.linalg.norm(rel_pos)

            # 4. Calculate Elevation (phi)
            # Angle from XZ plane to Y axis
            phi_rad = np.arcsin(y / (r + 1e-8))
            phi_deg = np.degrees(phi_rad)

            # 5. Calculate Azimuth (theta)
            # Angle in the XZ plane
            theta_rad = np.arctan2(z, x)
            theta_deg = np.degrees(theta_rad)

            # 6. Append all 4 parameters
            # face_idx is stored as a float so the whole array is homogeneous
            vec.extend([r, theta_deg, phi_deg])

        return np.array(vec)
