import type {
  Camera,
  CameraSaveEvent,
} from "~/messages/protobufs/autosave_event";
import {
  CameraEventType,
  CameraSaveEventSeries,
} from "~/messages/protobufs/autosave_event";
import type { ICamera } from "~/types/camera";
import type { SceneStates } from "~/types/scene-states";
import * as THREE from "three";

function isEqual(a: Camera, b: Camera): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export type AutosaveEvent =
  | { type: "delete"; data: string }
  | { type: "upsert"; data: ICamera };

export function transformCameraToProtoEvent(cam: ICamera): Omit<Camera, "id"> {
  const quaternion = new THREE.Quaternion();
  quaternion.setFromEuler(cam.rotation);
  return {
    name: cam.name,
    angleX: quaternion.x,
    angleY: quaternion.y,
    angleZ: quaternion.z,
    angleW: quaternion.w,
    posX: cam.position.x,
    posY: cam.position.y,
    posZ: cam.position.z,
    fov: cam.fov,
    frustumColor: cam.frustumColor,
    frustumLength: cam.frustumLength,
    isHidingArrows: cam.isHidingArrows,
    isHidingWheels: cam.isHidingWheels,
    isLockingPosition: cam.isLockingPosition,
    isLockingRotation: cam.isLockingRotation,
    isHidingFrustum: cam.isHidingFrustum,
  };
}

export function transformCameraToProtoEventWithId(
  camId: string,
  cam: ICamera,
): Camera {
  return { ...transformCameraToProtoEvent(cam), id: camId };
}

export function useAutosave(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  if (workspace == null) {
    return;
  }

  const lastSynced: Map<string, Camera> = new Map(
    Object.entries(sceneStates.cameras!).map(([camId, cam]) => {
      return [camId, transformCameraToProtoEventWithId(camId, cam)];
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
      const formattedCam = transformCameraToProtoEventWithId(camId, cam);
      // If is newly added, or changed
      if (prev == undefined || !isEqual(prev, formattedCam)) {
        changed.push({
          type: CameraEventType.CAMERA_EVENT_TYPE_UPSERT,
          upsert: formattedCam,
        });
        lastSynced.set(camId, transformCameraToProtoEventWithId(camId, cam));
      }
    }

    if (changed.length > 0 && sceneStates.websocket != undefined) {
      sceneStates.websocket.send(
        CameraSaveEventSeries.encode({ events: changed }).finish().buffer,
      );
    }

    sceneStates.markedForCheck.clear();
  }, 2000);
}
