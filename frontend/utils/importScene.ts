import type { ICamera } from "~/types/camera";
import type { Camera } from "~/messages/protobufs/autosave_event";
import type { SceneStates } from "~/types/scene-states";
import { transformProtoEventToCamera } from "~/components/3d/scene-states-provider/create-scene-states";

export function importJsonToCameras(
  sceneStates: SceneStates,
  jsonData: string,
) {
  try {
    const sceneCameras = sceneStates.cameras;
    const data = JSON.parse(jsonData) as Record<string, Camera>;

    const newCameras: Record<string, ICamera> = {};

    for (const [id, camera] of Object.entries(data)) {
      newCameras[id] = transformProtoEventToCamera(camera);
    }

    Object.keys(sceneCameras).forEach((key) => {
      Reflect.deleteProperty(sceneCameras, key);
    });

    Object.assign(sceneCameras, newCameras);

    for (const camId of Object.keys(newCameras)) {
      sceneStates.markedForCheck.add(camId);
    }
  } catch (err) {
    console.error("Failed to import cameras JSON:", err);
  }
}
