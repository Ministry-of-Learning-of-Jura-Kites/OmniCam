import sys
import os

sys.path.append(os.path.abspath(os.path.join("..")))

import asyncio
import json
import math
from os import path
import time
from typing import Any, Dict, List, Tuple
import uuid
from pydantic import BaseModel, ValidationError
import redis.asyncio as redis
from scipy.spatial.distance import cdist
from src.cost_functions import total_cost
from google.protobuf.json_format import MessageToJson
import messages.protobufs.camera_pb2 as cam_pb
import messages.protobufs.optimization_pb2 as opt_pb
from algorithms.differential_evolution import optimize_de
import numpy as np
from state import CameraConfiguration, CameraState, State
import quaternion
from utils import (
    center_of_face,
    look_at_quaternion,
    normal_vec_of_face,
)
from env import env_settings
import logging
import vtk
from dev.visualization import init_3d_scene, render_from_state
import pyvista as pv
from basic_types import Array4x3
from main import assign_faces

pl = None
if env_settings.dev_mode:
    from pyvistaqt import BackgroundPlotter

    pl = BackgroundPlotter()

gltf = (
    pv.read(path.join("/home/frook/Downloads/omnicam/test case a.glb"))
    .combine()
    .extract_surface()
    .triangulate()
    .clean()
)

gltf_locator = vtk.vtkStaticCellLocator()
gltf_locator.SetDataSet(gltf)
gltf_locator.BuildLocator()

face = np.array([[2, 2 + 10, 0], [-2, 2 + 10, 0], [-2, -2 + 10, 0], [2, -2 + 10, 0]])
cam_config = CameraConfiguration(
    pixels=[5000, 5000],
    vfov=60,
    name="f",
)

state = State(
    faces=[face],
    face_to_cam=dict(),
    face_centers=[center_of_face(face)],
    cameras=[
        CameraState(
            faces=None,
            pos=[5, 0, 0],
            angle=quaternion.from_vector_part([0, 0, 0, 1]),
            center_of_faces=None,
            camera_config=cam_config,
            name="gg",
        )
    ],
    scale=1,
    gltf=gltf,
    gltf_locator=gltf_locator,
)

seed = 2000

num_faces = len(state.faces)
num_cameras = len(state.cameras)

state = assign_faces(state, seed)

if env_settings.dev_mode:
    init_3d_scene(pl, state)
    render_from_state(pl, state)
    pl.show()

start_time = time.perf_counter()
# from cost_functions import total_cost
# print(total_cost(state))
# breakpoint()

# final_state = optimize_pso(
#     state,
#     # pl,
#     None,
# )
final_state = optimize_de(state, seed)

end_time = time.perf_counter()
elapsed_time = end_time - start_time
print(f"Elapsed time: {elapsed_time:.4f} seconds")

if env_settings.dev_mode:
    render_from_state(pl, final_state)
    breakpoint()

total = total_cost(final_state, True)
print("total cost: ", total)

if env_settings.dev_mode:
    pl.close()
