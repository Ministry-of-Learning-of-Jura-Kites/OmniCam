import json
import time
import numpy as np
from main import OptimizeRequest, optimize
import uuid
import base64


small_model_id_raw = "/home/frook/Downloads/living 1.14"
medium_model_id_raw = "/home/frook/Downloads/living 5.57"
large_model_id_raw = "/home/frook/Downloads/living"

times_avg = []
for id in [small_model_id_raw, medium_model_id_raw, large_model_id_raw]:
    times = []
    for i in range(10):
        with open("src/evaluation/ff.json", "r") as file:
            req = json.loads(file.read())
            data = json.loads(req["data"])
            data["model_id"] = str(id)

        payload = OptimizeRequest.model_validate(data)

        start = time.perf_counter()
        optimize(payload, 2000 + i)
        end = time.perf_counter()

        times.append(end - start)
    times_avg.append(np.array(times).mean())

print(times_avg)
