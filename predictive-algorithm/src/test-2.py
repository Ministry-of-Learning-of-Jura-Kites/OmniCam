import pyvista as pv
import quaternion
import vtk
from state import CameraConfiguration, CameraState, State
import numpy as np
from scipy.spatial.transform import Rotation as R

# from cost_functions.occlusion_cost import generate_raycast_depth_map

gltf = (
    pv.read("~/Downloads/omnicam/oxygenai-office-hallway.glb")
    .combine()
    .extract_surface()
    .triangulate()
    .clean()
)


print(f"Model Center: {gltf.center}")

# Print the bounds: (xmin, xmax, ymin, ymax, zmin, zmax)
print(f"Model Bounds: {gltf.bounds}")

gltf_locator = vtk.vtkStaticCellLocator()
gltf_locator.SetDataSet(gltf)
gltf_locator.BuildLocator()

faces = []
cam_config = CameraConfiguration(
    pixels=[6560, 3100],
    vfov=40.15,
    name="ggg",
)
cameras = [
    CameraState(
        faces=None,
        pos=np.array([-0.770839711791415, 1.2508888941895868, -2.446445983191274]),
        angle=quaternion.quaternion(
            0.04838919297746729,  # w
            -0.008805358656745012,
            -0.982652989534941,
            -0.17881290171561645,
        ),
        center_of_faces=None,
        camera_config=cam_config,
        name="gg",
    )
]


def get_rotated_basis(qua: quaternion):
    # cam_state.quaternion should be [x, y, z, w]
    w, x, y, z = quaternion.as_float_array(qua)
    q = [x, y, z, w]
    rotation = R.from_quat(q)

    # 1. Define the Native Basis (Three.js convention)
    native_forward = np.array([0, 0, -1])
    native_up = np.array([0, 1, 0])
    native_right = np.array([1, 0, 0])

    # 2. Rotate them into World Space
    forward = rotation.apply(native_forward)
    up = rotation.apply(native_up)
    right = rotation.apply(native_right)

    return forward, up, right


plotter = pv.Plotter()
plotter.add_mesh(gltf, color="white", show_edges=True)

# Visualize the Camera Position
plotter.add_points(
    cameras[0].pos,
    color="red",
    point_size=10,
    render_points_as_spheres=True,
)

# forward_vec, _, _ = get_rotated_basis(
#     [
#         0.005761262102837728,  # w
#         -0.0006363073435811718,
#         0.993939400830999,
#         0.10977645670935873,
#     ]
# )
# plotter.add_arrows(cameras[0].pos, forward_vec, mag=2, color="blue")

# plotter.add_axes()
# plotter.show()

state = State(
    faces=faces,
    face_to_cam=dict(),
    face_centers=list(),
    cameras=cameras,
    scale=1,
    gltf=gltf,
    gltf_locator=gltf_locator,
)

from PIL import Image

import numpy as np
import vtk
from tqdm import tqdm


def generate_raycast_depth_map(
    state, cam_state, near, far, resolution=(100, 100), log_interval=10
):
    """
    Generates a depth map via ray casting with progress logging.
    """
    h_res, v_res = resolution
    depth_map = np.zeros((v_res, h_res))

    pos = np.array(cam_state.pos)

    # Standard Pinhole Camera Vectors (assuming identity rotation for now)
    # Note: In a real app, these should be transformed by cam_state.rotation_matrix
    forward, up, right = get_rotated_basis(cam_state.angle)

    half_fov_rad = np.deg2rad(cam_state.camera_config.vfov) / 2.0
    h_far = np.tan(half_fov_rad)
    aspect_ratio = h_res / v_res
    w_far = h_far * aspect_ratio

    # Pre-allocate VTK objects to reuse (Huge speed boost)
    points = vtk.vtkPoints()
    cell_ids = vtk.vtkIdList()

    print(f"Starting raycast at {h_res}x{v_res}...")

    # Using tqdm for a visual progress bar
    for v in tqdm(range(v_res), desc="Rendering Depth Map"):
        # Map v once per row
        ndc_v = 1.0 - 2.0 * (v + 0.5) / v_res

        for u in range(h_res):
            # NDC u: left is -1, right is 1
            ndc_u = 2.0 * (u + 0.5) / h_res - 1.0

            # RAY DIRECTION:
            # Start at 'forward', then offset by the Right and Up vectors
            direction = forward + (right * ndc_u * w_far) + (up * ndc_v * h_far)
            ray_dir = direction / np.linalg.norm(direction)

            p_start = pos
            p_end = pos + ray_dir * 1000.0

            # Reset VTK output containers
            points.Reset()
            cell_ids.Reset()

            # Cast ray
            hit = state.gltf_locator.IntersectWithLine(
                p_start, p_end, 0.0001, points, cell_ids
            )

            if hit > 0:
                hit_point = np.array(points.GetPoint(0))
                depth = np.linalg.norm(hit_point - pos)
                depth_map[v, u] = (depth - near) / (far - near)
            else:
                depth_map[v, u] = 0

    return depth_map


def convert_to_depth_with_white_bg(depth_map):
    """
    Near = White (255)
    Far (1000) = Black (0)
    Infinite/Miss = White (255)
    """

    # 3. Invert so Near is 255 and Far (1000) is 0
    inverted_depth = depth_map * 255

    return inverted_depth.astype(np.uint8)


# --- Execution ---
NEAR = 0.1
FAR = 20.0

# Generate the map
raw_map = generate_raycast_depth_map(
    state,
    cameras[0],
    near=NEAR,
    far=FAR,
    resolution=(1532, 811),
    # resolution=(500, 500),
)

# Process with white background
final_image_array = convert_to_depth_with_white_bg(raw_map)

# Save/Show
im = Image.fromarray(final_image_array, "L")
im.save("/home/frook/Downloads/gg1.png")
