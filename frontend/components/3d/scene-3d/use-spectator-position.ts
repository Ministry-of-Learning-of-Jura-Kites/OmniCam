import { SPECTATOR_MOVING_SENTIVITY } from "~/constants";
import * as THREE from "three";
import type { SceneStates } from "~/types/scene-states";

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
    lastKeyDown = Date.now();
    if (sceneStates.tresContext.value?.camera == undefined) {
      setTimeout(() => requestAnimationFrame(refreshCameraState), 10);
      return;
    }
    const spectatorCamera = sceneStates.tresContext.value?.camera;
    for (const [key, isDown] of Object.entries(isKeyDown) as [
      FunctionalityKey,
      boolean,
    ][]) {
      if (!isDown) {
        continue;
      }
      const forward = new THREE.Vector3();
      spectatorCamera.getWorldDirection(forward);
      const up = new THREE.Vector3();
      up.copy(spectatorCamera.up).applyQuaternion(spectatorCamera.quaternion);
      const right = new THREE.Vector3();
      right.crossVectors(forward, up).normalize();
      let deltaVec = new THREE.Vector3();
      switch (key) {
        case "KeyW":
          deltaVec = forward.multiplyScalar(
            SPECTATOR_MOVING_SENTIVITY * duration,
          );
          break;
        case "KeyS":
          deltaVec = forward.multiplyScalar(
            -SPECTATOR_MOVING_SENTIVITY * duration,
          );
          break;
        case "KeyA":
          deltaVec = right.multiplyScalar(
            -SPECTATOR_MOVING_SENTIVITY * duration,
          );
          break;
        case "KeyD":
          deltaVec = right.multiplyScalar(
            SPECTATOR_MOVING_SENTIVITY * duration,
          );
          break;
        case "Space":
          deltaVec.y = SPECTATOR_MOVING_SENTIVITY * duration;
          break;
        case "Shift":
          deltaVec.y = -SPECTATOR_MOVING_SENTIVITY * duration;
          break;
        default:
          break;
      }
      sceneStates.currentCameraPosition.value?.add(deltaVec);
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
