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

function isEqual(a: Camera, b: Camera): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export type AutosaveEvent =
  | { type: "delete"; data: string }
  | { type: "upsert"; data: ICamera };

function formatCam(camId: string, cam: ICamera): Camera {
  const quaternion = new THREE.Quaternion();
  quaternion.setFromEuler(cam.rotation);
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
    fov: cam.fov,
    isHidingArrows: cam.isHidingArrows,
    isHidingWheels: cam.isHidingWheels,
  };
}

export function useAutosave(sceneStates: SceneStates) {
  const lastSynced: Map<string, Camera> = new Map(
    Object.entries(sceneStates.cameras!).map(([camId, cam]) => {
      return [camId, formatCam(camId, cam)];
    }),
  );

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
      const formattedCam = formatCam(camId, cam);
      // If is newly added, or changed
      if (prev == undefined || !isEqual(prev, formattedCam)) {
        changed.push({
          type: CameraEventType.CAMERA_EVENT_TYPE_UPSERT,
          upsert: formattedCam,
        });
        lastSynced.set(camId, formatCam(camId, cam));
      }
    }

    if (changed.length > 0) {
      sceneStates.websocket.send(
        CameraSaveEventSeries.encode({ events: changed }).finish().buffer,
      );
    }

    sceneStates.markedForCheck.clear();
  }, 2000);
}
