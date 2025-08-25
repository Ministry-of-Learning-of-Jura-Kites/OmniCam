import { camera } from "./camera";

import * as THREE from "three";

const functionalityKeys = ["a", "w", "s", "d", " ", "Shift"] as const;

type FunctionalityKey = (typeof functionalityKeys)[number];

const isKeyDown: Record<FunctionalityKey, boolean> = functionalityKeys.reduce(
  (acc, key) => {
    acc[key] = false;
    return acc;
  },
  {} as Record<FunctionalityKey, boolean>,
);

function isFunctionalityKey(key: string): key is FunctionalityKey {
  return (functionalityKeys as readonly string[]).includes(key);
}

function onKeyDown(e: KeyboardEvent) {
  if (camera?.value == undefined || e.repeat) {
    return;
  }
  if (isFunctionalityKey(e.key)) {
    isKeyDown[e.key] = true;
  }
}

function onKeyUp(e: KeyboardEvent) {
  if (camera?.value == undefined) {
    return;
  }
  if (isFunctionalityKey(e.key)) {
    isKeyDown[e.key] = false;
  }
}

function setup() {
  setInterval(() => {
    if (camera?.value == undefined) {
      return;
    }
    for (const [key, isDown] of Object.entries(isKeyDown) as [
      FunctionalityKey,
      boolean,
    ][]) {
      if (!isDown) {
        continue;
      }
      const forward = new THREE.Vector3();
      camera.value.getWorldDirection(forward);

      const up = new THREE.Vector3();
      up.copy(camera.value.up).applyQuaternion(camera.value.quaternion);

      const right = new THREE.Vector3();
      right.crossVectors(forward, up).normalize();

      let deltaVec = new THREE.Vector3();

      switch (key) {
        case "w":
          deltaVec = forward.multiplyScalar(0.1);
          break;
        case "s":
          deltaVec = forward.multiplyScalar(-0.1);
          break;
        case "a":
          deltaVec = right.multiplyScalar(-0.1);
          break;
        case "d":
          deltaVec = right.multiplyScalar(0.1);
          break;
        case " ":
          deltaVec.y = 0.1;
          break;
        case "Shift":
          deltaVec.y = -0.1;
          break;
        default:
          break;
      }
      camera.value.position.add(deltaVec);
    }
  }, 10);
}

export const SpectatorPosition = {
  onKeyUp,
  onKeyDown,
  setup,
};
