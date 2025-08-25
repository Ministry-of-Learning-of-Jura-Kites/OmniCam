import { SPECTATOR_MOVING_SENTIVITY } from "~/constants";
import { camera, cameraPosition } from "./refs";

import * as THREE from "three";
import { ca } from "zod/v4/locales";

const functionalityKeys = [
  "KeyA",
  "KeyW",
  "KeyS",
  "KeyD",
  "Space",
  "Shift",
] as const;

type FunctionalityKey = (typeof functionalityKeys)[number];

let isKeyDown: Partial<Record<FunctionalityKey, boolean>> = {};

function isFunctionalityKey(key: string): key is FunctionalityKey {
  return (functionalityKeys as readonly string[]).includes(key);
}

function onKeyDown(e: KeyboardEvent) {
  if (camera?.value == undefined || e.repeat) {
    return;
  }

  // console.log("press", e.code);
  if (e.code == "ShiftLeft" || e.code == "ShiftRight") {
    isKeyDown["Shift"] = true;
  }
  if (isFunctionalityKey(e.code)) {
    isKeyDown[e.code] = true;
  }
}

function onKeyUp(e: KeyboardEvent) {
  if (camera?.value == undefined) {
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
  if (camera?.value == undefined) {
    requestAnimationFrame(update);
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
    cameraPosition.add(deltaVec);
  }
  requestAnimationFrame(update);
}

function onBlur(_e: FocusEvent) {
  isKeyDown = {};
}

export const SpectatorPosition = {
  onKeyUp,
  onKeyDown,
  update,
  onBlur,
};
