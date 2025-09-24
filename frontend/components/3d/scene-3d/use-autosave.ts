import type {
  Camera,
  CameraSaveEvent,
} from "~/messages/protobufs/protobufs/autosave_event";
import {
  CameraEventType,
  CameraSaveEventSeries,
} from "~/messages/protobufs/protobufs/autosave_event";
import type { ICamera } from "~/types/camera";
import type { SceneStates } from "~/types/scene-states";
import * as THREE from "three";

function isEqual(a: ICamera, b: ICamera): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export type AutosaveEvent =
  | { type: "delete"; data: string }
  | { type: "upsert"; data: ICamera };

function formatCam(camId: string, cam: ICamera): Camera {
  const quaternion = new THREE.Quaternion();
  cam.rotation.setFromQuaternion(quaternion);
  return {
    id: camId,
    name: cam.name,
    angleX: quaternion.x,
    angleY: quaternion.y,
    angleZ: quaternion.z,
    angleW: quaternion.w,
    posX: cam.position.x,
    posY: cam.position.y,
    posZ: cam.position.z,
  };
}

export function useAutosave(sceneStates: SceneStates) {
  const lastSynced: Map<string, ICamera> = new Map(
    Object.entries(sceneStates.cameras!),
  );

  // detect changes batchly
  // watch(
  //   () => sceneStates.cameras,
  //   (newVal) => {
  //     for (const [camId, cam] of Object.entries(newVal)) {
  //       const prevCam = oldVal[camId];
  //       console.log(cam, prevCam, "gg");
  //       if (prevCam == undefined || !isEqual(prevCam, cam)) {
  //         console.log("ggfff");
  //         markedForCheck.add(camId);
  //       }
  //     }
  //   },
  //   { deep: true },
  // );

  setInterval(() => {
    if (sceneStates.markedForCheck.size == 0) {
      return;
    }

    const newVal = sceneStates.cameras;
    const changed: CameraSaveEvent[] = [];

    // check new or updated
    for (const camId of sceneStates.markedForCheck) {
      const prev = lastSynced.get(camId);
      const cam = newVal[camId];
      if (cam == undefined && prev == undefined) {
        continue;
      }
      // If is deleted
      if (cam == undefined) {
        lastSynced.delete(camId);
        changed.push({
          type: CameraEventType.CAMERA_EVENT_TYPE_DELETE,
          deleteId: camId,
        });
        continue;
      }
      // If is newly added, or changed
      if (prev == undefined || !isEqual(prev, cam)) {
        changed.push({
          type: CameraEventType.CAMERA_EVENT_TYPE_UPSERT,
          upsert: formatCam(camId, cam),
        });
        lastSynced.set(camId, structuredClone(toRaw(cam)));
      }
    }

    console.log("changed", changed);

    if (changed.length > 0) {
      sceneStates.websocket.send(
        CameraSaveEventSeries.encode({ events: changed }).finish().buffer,
      );
    }

    sceneStates.markedForCheck.clear();
  }, 2000);
}
