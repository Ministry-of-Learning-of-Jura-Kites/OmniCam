import numpy as np
from state import State
import quaternion
from utils import get_seeded_color_rgb
from env import env_settings
import vtk
import pyvista as pv

if env_settings.dev_mode:
    import pyvistaqt


def init_3d_scene(pl: "pyvistaqt.BackgroundPlotter | None", state: State):
    if not env_settings.dev_mode:
        return

    if pl is None:
        return
    pl.add_mesh(state.gltf)
    pl.show_grid(color="gray", location="outer")
    face_mesh: pv.PolyData | None = None
    for i, face in enumerate(state.faces):
        # face_center = center_of_face(face)

        faces = np.hstack([[4, 0, 1, 2, 3]])
        face_mesh = pv.PolyData(face, faces=faces)
        color = get_seeded_color_rgb(state.face_to_cam[i])
        pl.add_mesh(face_mesh, color=color)

    for camera in state.cameras:
        # camera.angle = look_at_quaternion(face_center - camera.pos)

        arrow = pv.Arrow(start=(0, 0, 0), direction=(1.0, 0.0, 0.0))
        camera.meshes.camera_actor = pl.add_mesh(arrow, color=color)
        silhouette_actor = pl.add_silhouette(
            arrow,
            color="white",
            line_width=8.0,
        )
        camera.meshes.camera_silhouette_actor = silhouette_actor

        temp_cam = pv.Camera()
        temp_cam.position = np.array([0, 0, 0])
        temp_cam.clipping_range = (0.1, 10.0)
        temp_cam.focal_point = temp_cam.position + np.array([1, 0, 0])
        temp_cam.up = (0, 1, 0)
        temp_cam.view_angle = camera.camera_config.vfov
        aspect = camera.camera_config.pixels[0] / camera.camera_config.pixels[1]
        frustum = temp_cam.view_frustum(aspect)
        camera.meshes.frustum_actor = pl.add_mesh(
            frustum, color=color, style="wireframe", opacity=0.5, line_width=2
        )

    pl.camera.up = (0, 1, 0)
    pl.camera_set = True
    pl.add_axes()
    pl.show_axes()
    pl.reset_camera(render=True, bounds=face_mesh.bounds)
    pl.enable_trackball_style()


def render_from_state(pl: pv.Plotter, state: State):
    if not env_settings.dev_mode:
        return

    # pl.clear()
    for i, camera in enumerate(state.cameras):
        rot_mat = quaternion.as_rotation_matrix(camera.angle)

        transform = np.eye(4)
        transform[:3, :3] = rot_mat
        transform[:3, 3] = camera.pos
        vtk_matrix = vtk.vtkMatrix4x4()
        for row in range(4):
            for col in range(4):
                vtk_matrix.SetElement(row, col, transform[row, col])
        camera.meshes.camera_actor.SetUserMatrix(vtk_matrix)
        camera.meshes.camera_silhouette_actor.SetUserMatrix(vtk_matrix)
        camera.meshes.frustum_actor.SetUserMatrix(vtk_matrix)
