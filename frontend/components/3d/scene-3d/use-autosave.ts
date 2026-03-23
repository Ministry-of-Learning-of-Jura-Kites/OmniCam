import {
  type Camera,
  type AutosaveEvent,
  type CoverageFace,
  AutosaveMessage,
  AutosaveResponse,
} from "~/messages/protobufs/autosave_event";
import type { ICamera } from "~/types/camera";
import type { SceneStates } from "~/types/scene-states";
import { Quaternion } from "three";
import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";

function isEqual<T>(a: T, b: T): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export function transformCameraToProtoEvent(cam: ICamera): Omit<Camera, "id"> {
  const quaternion = new Quaternion();
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
    widthRes: cam.widthRes,
    heightRes: cam.heightRes,
    fov: cam.fov,
    frustumColor: cam.frustumColor,
    frustumLength: cam.frustumLength,
    isHidingArrows: cam.isHidingArrows,
    isHidingWheels: cam.isHidingWheels,
    isLockingPosition: cam.isLockingPosition,
    isLockingRotation: cam.isLockingRotation,
    isHidingFrustum: cam.isHidingFrustum,
    distortion: structuredClone(toRaw(cam.distortion)),
  };
}

export function transformCameraToProtoEventWithId(
  camId: string,
  cam: ICamera,
): Camera {
  return { ...transformCameraToProtoEvent(cam), id: camId };
}

export function transformFaceToProto(
  id: string,
  face: ProcessedCoverageFace,
): CoverageFace {
  return {
    id: id,
    points: face.points.map(numbersToVector3),
    color: face.color,
    hidden: face.hidden,
  };
}

export function useAutosave(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  if (workspace == null) return;

  const lastSyncedCams: Map<string, Camera> = new Map(
    Object.entries(sceneStates.cameras!).map(([camId, cam]) => [
      camId,
      transformCameraToProtoEventWithId(camId, cam),
    ]),
  );

  const isServerUpdate = ref(false);

  const lastSyncedFaces: Map<string, CoverageFace> = new Map(
    Object.entries(sceneStates.facesManagement.faces).map(([id, face]) => [
      id,
      transformFaceToProto(id, face),
    ]),
  );

  const markedFacesForCheck = new Set<string>();

  watch(
    () => sceneStates.facesManagement.faces,
    (newFaces, oldFaces) => {
      // Track additions and updates
      Object.keys(newFaces).forEach((id) => markedFacesForCheck.add(id));
      // Track deletions (keys present in old but not in new)
      if (oldFaces) {
        Object.keys(oldFaces).forEach((id) => {
          if (!newFaces[id]) markedFacesForCheck.add(id);
        });
      }
    },
    { deep: true },
  );

  onMounted(() => {
    watch(
      () => sceneStates.websocket?.data.value,
      async (messageBlob) => {
        if (!messageBlob) return;
        const buf = await (messageBlob as Blob).arrayBuffer();
        const resp = AutosaveResponse.decode(new Uint8Array(buf));

        sceneStates.lastSyncedVersion.value = resp.lastUpdatedVersion;

        // if (resp.calibrationAck) {
        //   isServerUpdate.value = true;
        //   sceneStates.calibrationScale.value = resp.calibrationAck.scaleFactor;
        //   sceneStates.calibrationHeight.value = resp.calibrationAck.modelHeight;
        //   sceneStates.calibrationVersion.value =
        //     resp.calibrationAck.lastUpdatedVersion;
        //   await nextTick();
        //   isServerUpdate.value = false;
        // }
      },
    );

    watch(
      () => [
        sceneStates.calibration.scale,
        sceneStates.calibration.heightOffset,
      ],
      ([newScale, newHeight], [oldScale, oldHeight]) => {
        if (isServerUpdate.value) return;
        if (newScale !== oldScale || newHeight !== oldHeight) {
          sceneStates.calibration.dirty = true;
        }
      },
    );

    function updateCams(changed: AutosaveEvent[]) {
      if (!sceneStates.markedForCheck.value) {
        return;
      }

      // Cameras
      if (sceneStates.markedForCheck.value) {
        for (const [camId, cam] of Object.entries(sceneStates.cameras)) {
          const prev = lastSyncedCams.get(camId);
          const formattedCam = transformCameraToProtoEventWithId(camId, cam);

          if (prev == undefined || !isEqual(prev, formattedCam)) {
            changed.push({ upsert: { camera: formattedCam } });
            lastSyncedCams.set(camId, formattedCam);
          }
        }

        // Check for deleted cameras
        for (const camId of lastSyncedCams.keys()) {
          if (!sceneStates.cameras[camId]) {
            lastSyncedCams.delete(camId);
            changed.push({ delete: { id: camId } });
          }
        }
      }

      sceneStates.markedForCheck.value = false;
    }

    function updateCalibration(changed: AutosaveEvent[]) {
      if (sceneStates.calibration.dirty) {
        changed.push({
          calibrate: {
            scaleFactor: sceneStates.calibration.scale,
            modelHeight: sceneStates.calibration.heightOffset,
          },
        });
        sceneStates.calibration.dirty = false;
      }
    }

    function updateFaces(changed: AutosaveEvent[]) {
      if (markedFacesForCheck.size == 0) {
        return;
      }
      for (const faceId of markedFacesForCheck) {
        const prev = lastSyncedFaces.get(faceId);
        const face = sceneStates.facesManagement.faces[faceId];

        // 1. Handle Deletion
        if (face === undefined) {
          if (prev !== undefined) {
            lastSyncedFaces.delete(faceId);
            changed.push({ facesDelete: { id: faceId } });
          }
          continue;
        }

        // 2. Handle Upsert (New or Changed)
        const formattedFace = transformFaceToProto(faceId, face);

        // Use isEqual (from lodash or similar) to avoid redundant network traffic
        if (prev === undefined || !isEqual(prev, formattedFace)) {
          changed.push({ facesUpsert: { coverageFace: formattedFace } });
          lastSyncedFaces.set(faceId, formattedFace);
        }
      }
      markedFacesForCheck.clear();
    }

    setInterval(() => {
      if (!sceneStates.websocket) return;

      const changed: AutosaveEvent[] = [];

      updateCams(changed);

      updateCalibration(changed);

      updateFaces(changed);

      if (changed.length > 0) {
        sceneStates.localVersion.value += 1;
        const encoded = AutosaveMessage.encode({
          version: sceneStates.localVersion.value,
          events: changed,
        }).finish();
        sceneStates.websocket.send(encoded.buffer);
      }
    }, 2000);
  });

  // function requestOptimize(
  //   faceIds: string[],
  //   cameras: { cam: ICamera; amount: number }[],
  // ) {
  //   sceneStates.facesManagement.faces.value;
  // }
}
