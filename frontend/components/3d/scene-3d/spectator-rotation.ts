import { SPECTATOR_ROTAING_SENTIVITY } from "~/constants";
import {
  currentCameraRotation,
  isDraggingObject,
  tresCanvasParent,
} from "./refs";

const isDragging = ref(false);

const maxPitch = Math.PI / 2 - 0.01;
const minPitch = -Math.PI / 2 + 0.01;

async function onPointerDown(_e: PointerEvent) {
  isDragging.value = true;

  document.addEventListener("pointermove", onPointerMove);
  document.addEventListener("pointerup", onPointerUp);
  await tresCanvasParent.value?.requestPointerLock();
}

function onPointerUp(_e: PointerEvent) {
  document.exitPointerLock();
  isDragging.value = false;
  document.removeEventListener("pointermove", onPointerMove);
  document.removeEventListener("pointerup", onPointerUp);
}

function normalizeAngle(rad: number) {
  return ((rad % (2 * Math.PI)) + 2 * Math.PI) % (2 * Math.PI);
}

function onPointerMove(e: PointerEvent) {
  if (tresCanvasParent?.value == undefined || isDraggingObject.value) {
    return;
  }

  if (!isDragging.value || currentCameraRotation.value == null) return;

  const deltaX = e.movementX;
  let yaw =
    currentCameraRotation.value.y - deltaX * SPECTATOR_ROTAING_SENTIVITY;
  yaw = normalizeAngle(yaw);
  currentCameraRotation.value.y = yaw;

  const deltaY = e.movementY;
  let pitch =
    currentCameraRotation.value.x - deltaY * SPECTATOR_ROTAING_SENTIVITY;
  pitch = Math.max(minPitch, Math.min(maxPitch, pitch));
  currentCameraRotation.value.x = pitch;
}

function onBlur(_e: FocusEvent) {
  document.exitPointerLock();
  isDragging.value = false;
}

export const SpectatorRotation = {
  onPointerDown,
  onBlur,
};
