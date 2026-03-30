{
    "data": '{"cam_configs":[{"amount":1,"name":"ARRI ALEXA 265 5.1K 1.65:1","pixels":[5120,3100],"vfov":40.15}],"faces":[[[4.320067888482462,0.20703368782832832,3.716043351018164],[4.096433048629512,0.2059052431749382,5.3215276718796565],[4.1315001071190425,2.334402403311326,5.333298168030135],[4.3528990316138625,2.3823986676166795,3.7496699281210577]]],"job_id":"7f54881d-7361-4427-8753-ba64106abda4","model_id":"6813a912-e364-4190-9041-e3a772dd6844","project_id":"2f310ec5-bd86-4230-8051-4e60346d56ae","scale":1}'
}

import json
from os import path
import time
from cost_functions import total_cost
from algorithms.differential_evolution import optimize_de
import numpy as np
from state import CameraConfiguration, CameraState, State
import quaternion
from utils import center_of_face
import vtk
import pyvista as pv
from main import OptimizeRequest, assign_faces, optimize
import uuid
import base64


def parse_uuid_base64(base64_str: str) -> uuid.UUID:
    """
    Direct Python port of the Go ParseUuidBase64 function.
    Takes a 22-character Base64 string and returns a UUID object.
    """
    # 1. Add back the missing '=' padding
    # Base64 strings must have a length divisible by 4
    padding = "=" * (4 - len(base64_str) % 4)
    padded_str = base64_str + padding

    # 2. Decode using the URL-safe alphabet (+ -> -, / -> _)
    decoded_bytes = base64.urlsafe_b64decode(padded_str)

    # 3. Convert the 16 raw bytes into a Python UUID object
    return uuid.UUID(bytes=decoded_bytes)


small_model_id_raw = "aBOpEuNkQZCQQeOnct1oRA"
medium_model_id_raw = "9pUzFr5iRWeIcu50urUsvQ"
large_model_id_raw = "It6QmaHGSHCaTtiGSS13mA"
small_model_id = parse_uuid_base64(small_model_id_raw)
medium_model_id = parse_uuid_base64(medium_model_id_raw)
large_model_id = parse_uuid_base64(large_model_id_raw)

times_avg = []
for id in [medium_model_id]:
    times = []
    for i in range(10):
        start = time.perf_counter()
        with open("src/evaluation/ff.json", "r") as file:
            req = json.loads(file.read())
            data = json.loads(req["data"])
            data["model_id"] = str(id)

        payload = OptimizeRequest.model_validate(data)

        optimize(payload)
        end = time.perf_counter()

        times.append(end - start)
    times_avg.append(np.array(times).mean())

print(times_avg)
