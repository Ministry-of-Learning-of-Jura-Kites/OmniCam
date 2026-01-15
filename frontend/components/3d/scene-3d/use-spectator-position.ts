import { BASE_SENSITIVITY, SPECTATOR_MOVING_SENTIVITY } from "~/constants";
import { Vector3 } from "three";
import type { SceneStates } from "~/types/scene-states";
import { useSensitivity } from "#imports";
const functionalityKeys = [
  "KeyA",
  "KeyW",
  "KeyS",
  "KeyD",
  "Space",
  "Shift",
] as const;

type FunctionalityKey = (typeof functionalityKeys)[number];

export function useSpectatorPosition(sceneStates: SceneStates) {
  let isKeyDown: Partial<Record<FunctionalityKey, boolean>> = {};
  let lastKeyDown = Date.now();

  function isFunctionalityKey(key: string): key is FunctionalityKey {
    return (functionalityKeys as readonly string[]).includes(key);
  }

  function onKeyDown(e: KeyboardEvent) {
    if (sceneStates.tresContext.value?.camera == undefined || e.repeat) {
      return;
    }

    sceneStates.tresCanvasParent?.value?.focus();

    // console.log("press", e.code);
    if (e.code == "ShiftLeft" || e.code == "ShiftRight") {
      isKeyDown["Shift"] = true;
    }
    if (isFunctionalityKey(e.code)) {
      isKeyDown[e.code] = true;
    }
  }

  function onKeyUp(e: KeyboardEvent) {
    if (sceneStates.tresContext.value?.camera == undefined) {
      return;
    }

    // console.log("release", e.code);
    if (e.code == "ShiftLeft" || e.code == "ShiftRight") {
      isKeyDown["Shift"] = false;
    }
    if (isFunctionalityKey(e.code)) {
      isKeyDown[e.code] = false;
    }
  }

  function refreshCameraState() {
    const duration = Date.now() - lastKeyDown;
    const userSensitivity = useSensitivity();
    lastKeyDown = Date.now();
    if (sceneStates.tresContext.value?.camera == undefined) {
      setTimeout(() => requestAnimationFrame(refreshCameraState), 10);
      return;
    }
    const spectatorCamera = sceneStates.tresContext.value?.camera;
    const camPreviewId = sceneStates.currentCamId.value;
    const isLock = sceneStates.cameras[camPreviewId!]?.isLockingPosition;
    for (const [key, isDown] of Object.entries(isKeyDown) as [
      FunctionalityKey,
      boolean,
    ][]) {
      if (!isDown || isLock) {
        continue;
      }
      const forward = new Vector3();
      spectatorCamera.getWorldDirection(forward);
      const up = new Vector3();
      up.copy(spectatorCamera.up).applyQuaternion(spectatorCamera.quaternion);
      const right = new Vector3();
      right.crossVectors(forward, up).normalize();
      let deltaVec = new Vector3();

      switch (key) {
        case "KeyW":
          deltaVec = forward.multiplyScalar(
            SPECTATOR_MOVING_SENTIVITY *
              duration *
              userSensitivity.sensitivity.value.movement *
              BASE_SENSITIVITY,
          );
          break;
        case "KeyS":
          deltaVec = forward.multiplyScalar(
            -SPECTATOR_MOVING_SENTIVITY *
              duration *
              userSensitivity.sensitivity.value.movement *
              BASE_SENSITIVITY,
          );
          break;
        case "KeyA":
          deltaVec = right.multiplyScalar(
            -SPECTATOR_MOVING_SENTIVITY *
              duration *
              userSensitivity.sensitivity.value.movement *
              BASE_SENSITIVITY,
          );
          break;
        case "KeyD":
          deltaVec = right.multiplyScalar(
            SPECTATOR_MOVING_SENTIVITY *
              duration *
              userSensitivity.sensitivity.value.movement *
              BASE_SENSITIVITY,
          );
          break;
        case "Space":
          deltaVec.y =
            SPECTATOR_MOVING_SENTIVITY *
            duration *
            userSensitivity.sensitivity.value.movement *
            BASE_SENSITIVITY;
          break;
        case "Shift":
          deltaVec.y =
            -SPECTATOR_MOVING_SENTIVITY *
            duration *
            userSensitivity.sensitivity.value.movement *
            BASE_SENSITIVITY;
          break;
        default:
          break;
      }
      sceneStates.currentCam.value?.position.add(deltaVec);
    }
    setTimeout(() => requestAnimationFrame(refreshCameraState), 10);
  }

  function onBlur(_e: FocusEvent) {
    isKeyDown = {};
  }

  return {
    onKeyUp,
    onKeyDown,
    refreshCameraState,
    onBlur,
  };
}

export type SpectatorPosition = ReturnType<typeof useSpectatorPosition>;
