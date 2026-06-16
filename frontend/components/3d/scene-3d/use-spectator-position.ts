import { SPECTATOR_MOVING_SENTIVITY, SPECTATOR_SPEED_BOOST } from "~/constants";
import { Vector3 } from "three";
import type { SceneStates } from "~/types/scene-states";
import { useSensitivity } from "#imports";
const functionalityKeys = [
  "KeyA",
  "KeyW",
  "KeyS",
  "KeyD",
  "KeyX",
  "Space",
  "Shift",
  "Ctrl",
] as const;

type FunctionalityKey = (typeof functionalityKeys)[number];

export function useSpectatorPosition(
  sceneStates: SceneStates,
  workspace: string | null,
) {
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
    const spectatorCamera = sceneStates.tresContext.value?.camera.activeCamera;
    const camPreviewId = sceneStates.currentCamId.value;
    const camActive = sceneStates.currentCamId.value != null;
    const notInOwnWorkspace = workspace != "me";
    const isLock =
      sceneStates.cameras[camPreviewId!]?.isLockingPosition ||
      (camActive && notInOwnWorkspace);
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

      let multiplier =
        (SPECTATOR_MOVING_SENTIVITY *
          duration *
          userSensitivity.normalizedSensitivity.value.movement) /
        sceneStates.calibration.scale;

      if (isKeyDown.Shift) {
        multiplier *= SPECTATOR_SPEED_BOOST;
      }

      switch (key) {
        case "KeyW":
          deltaVec = forward.multiplyScalar(multiplier);
          break;
        case "KeyS":
          deltaVec = forward.multiplyScalar(-multiplier);
          break;
        case "KeyA":
          deltaVec = right.multiplyScalar(-multiplier);
          break;
        case "KeyD":
          deltaVec = right.multiplyScalar(multiplier);
          break;
        case "Space":
          deltaVec.y = multiplier;
          break;
        case "KeyX":
          deltaVec.y = -multiplier;
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

  onMounted(() => {
    refreshCameraState();
  });

  return {
    onKeyUp,
    onKeyDown,
    refreshCameraState,
    onBlur,
  };
}

export type SpectatorPosition = ReturnType<typeof useSpectatorPosition>;
