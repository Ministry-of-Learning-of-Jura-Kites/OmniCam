import numpy as np
from state import State
import quaternion
from utils import get_seeded_color_rgb
from env import env_settings
import vtk
import pyvista as pv
from cost_functions import total_cost

import vtk
import numpy as np
import quaternion

if env_settings.dev_mode:
    import pyvistaqt

import vtk
import numpy as np
import pyvista as pv
import quaternion

# Shared identifier for the UI element
COST_LABEL_NAME = "system_cost_overlay"


def update_system_visuals(pl: pv.Plotter, state: State, face_actors: list):
    """Shared helper to update cost text and face colors."""
    # 1. Update Cost Function
    cost = total_cost(state, True)

    # 2. Update Text Overlay (using shared name to overwrite)
    pl.add_text(
        f"SYSTEM COST: ${cost:,.2f}",
        name=COST_LABEL_NAME,
        position="upper_right",
        color="yellow",
        shadow=True,
        font_size=12,
    )

    # 3. Update Face Colors
    for i, actor in enumerate(face_actors):
        new_color = get_seeded_color_rgb(state.face_to_cam[i])
        actor.GetProperty().SetColor(new_color)

    pl.render()


def init_3d_scene(pl: "pyvistaqt.BackgroundPlotter | None", state: State):
    if not env_settings.dev_mode or pl is None:
        return

    pl.add_mesh(state.gltf, opacity=0.1)

    face_actors = []
    face_mesh = None
    for i, face in enumerate(state.faces):
        faces_conn = np.hstack([[4, 0, 1, 2, 3]])
        face_mesh = pv.PolyData(face, faces=faces_conn)
        actor = pl.add_mesh(
            face_mesh,
            color=get_seeded_color_rgb(state.face_to_cam[i]),
            name=f"face_{i}",
        )
        face_actors.append(actor)

    # Store face_actors on pl so render_from_state can find them
    pl.face_actors = face_actors

    def move_cam(idx, x=None, y=None, z=None, yaw=None, pitch=None, roll=0):
        if idx >= len(state.cameras):
            return print("Index out of bounds")

        cam = state.cameras[idx]
        if x is not None:
            cam.pos[0] = x
        if y is not None:
            cam.pos[1] = y
        if z is not None:
            cam.pos[2] = z

        if yaw is not None or pitch is not None:
            y_rad, p_rad, r_rad = (
                np.radians(yaw or 0),
                np.radians(pitch or 0),
                np.radians(roll),
            )
            cam.angle = quaternion.from_euler_angles(y_rad, p_rad, r_rad)

        rot_mat = quaternion.as_rotation_matrix(cam.angle)
        mat = np.eye(4)
        mat[:3, :3] = rot_mat
        mat[:3, 3] = cam.pos

        for attr in ["camera_actor", "frustum_actor", "camera_silhouette_actor"]:
            actor = getattr(cam.meshes, attr, None)
            if actor:
                actor.user_matrix = mat

        cam.forward_vec = rot_mat @ np.array([0, 0, -1])

        # Trigger shared visual update
        update_system_visuals(pl, state, face_actors)

    pl.move_cam = move_cam

    for i, camera in enumerate(state.cameras):
        color = get_seeded_color_rgb(state.face_to_cam[i])
        arrow = pv.Arrow(start=(0, 0, 0), direction=(0.0, 0.0, -1.0))

        camera.meshes.camera_actor = pl.add_mesh(
            arrow, color=color, name=f"cam_arrow_{i}"
        )
        camera.meshes.camera_silhouette_actor = pl.add_silhouette(
            arrow, color="white", line_width=4.0
        )

        temp_cam = pv.Camera()
        temp_cam.view_angle = camera.camera_config.vfov
        temp_cam.position = np.array([0, 0, 0])
        temp_cam.clipping_range = (0.1, 10.0)
        temp_cam.focal_point = temp_cam.position + np.array([0, 0, -1])
        temp_cam.up = (0, 1, 0)

        aspect = camera.camera_config.pixels[0] / camera.camera_config.pixels[1]
        frustum = temp_cam.view_frustum(aspect)
        camera.meshes.frustum_actor = pl.add_mesh(
            frustum, color=color, style="wireframe", opacity=0.5, name=f"cam_frust_{i}"
        )
        move_cam(i)

    pl.camera.up = (0, 1, 0)
    pl.add_axes()
    if face_mesh:
        pl.reset_camera(render=True, bounds=face_mesh.bounds)
    pl.enable_trackball_style()


def render_from_state(pl: pv.Plotter, state: State):
    if not env_settings.dev_mode:
        return

    for i, camera in enumerate(state.cameras):
        rot_mat = quaternion.as_rotation_matrix(camera.angle)
        transform = np.eye(4)
        transform[:3, :3] = rot_mat
        transform[:3, 3] = camera.pos

        vtk_matrix = vtk.vtkMatrix4x4()
        for row in range(4):
            for col in range(4):
                vtk_matrix.SetElement(row, col, transform[row, col])

        if camera.meshes.camera_actor:
            camera.meshes.camera_actor.SetUserMatrix(vtk_matrix)
        if camera.meshes.camera_silhouette_actor:
            camera.meshes.camera_silhouette_actor.SetUserMatrix(vtk_matrix)
        if camera.meshes.frustum_actor:
            camera.meshes.frustum_actor.SetUserMatrix(vtk_matrix)

        camera.forward_vec = rot_mat @ np.array([0, 0, -1])

    # Pull the actors list we saved in init_3d_scene
    face_actors = getattr(pl, "face_actors", [])
    update_system_visuals(pl, state, face_actors)


# Cam 0 moved to [1.45, 0.65, 0], Yaw: 90, Pitch: -37.5
# pl.move_cam(0,x=1.45,y=0.65,z=0,yaw=37.5,pitch=90)
