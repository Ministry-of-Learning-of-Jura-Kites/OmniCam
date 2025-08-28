import { SPECTATOR_MOVING_SENTIVITY } from "~/constants";
import * as THREE from "three";
import type { SceneStates } from "./use-scene-state";

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

  function isFunctionalityKey(key: string): key is FunctionalityKey {
    return (functionalityKeys as readonly string[]).includes(key);
  }

  function onKeyDown(e: KeyboardEvent) {
    e.preventDefault();
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

  function update() {
    if (sceneStates.tresContext.value?.camera == undefined) {
      requestAnimationFrame(update);
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
          deltaVec = forward.multiplyScalar(SPECTATOR_MOVING_SENTIVITY);
          break;
        case "KeyS":
          deltaVec = forward.multiplyScalar(-SPECTATOR_MOVING_SENTIVITY);
          break;
        case "KeyA":
          deltaVec = right.multiplyScalar(-SPECTATOR_MOVING_SENTIVITY);
          break;
        case "KeyD":
          deltaVec = right.multiplyScalar(SPECTATOR_MOVING_SENTIVITY);
          break;
        case "Space":
          deltaVec.y = SPECTATOR_MOVING_SENTIVITY;
          break;
        case "Shift":
          deltaVec.y = -SPECTATOR_MOVING_SENTIVITY;
          break;
        default:
          break;
      }
      sceneStates.currentCameraPosition.value?.add(deltaVec);
    }
    requestAnimationFrame(update);
  }

  function onBlur(_e: FocusEvent) {
    isKeyDown = {};
  }

  return {
    onKeyUp,
    onKeyDown,
    update,
    onBlur,
  };
}

export type SpectatorPosition = ReturnType<typeof useSpectatorPosition>;
