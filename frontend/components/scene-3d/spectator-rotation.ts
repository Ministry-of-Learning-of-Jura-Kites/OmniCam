import { SPECTATOR_ROTAING_SENTIVITY } from "~/constants";
import { camera, tresCanvasParent } from "./refs";

const isDragging = ref(false);

const maxPitch = Math.PI / 2 - 0.01;
const minPitch = -Math.PI / 2 + 0.01;

let yaw = 0;
let pitch = 0;

async function onPointerDown(e: PointerEvent) {
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

function onPointerMove(e: PointerEvent) {
  if (tresCanvasParent?.value == undefined) {
    return;
  }

  if (!isDragging.value) return;

  const deltaX = e.movementX;
  yaw -= deltaX * SPECTATOR_ROTAING_SENTIVITY;
  camera.value!.rotation.y = yaw;

  const deltaY = e.movementY;
  pitch -= deltaY * SPECTATOR_ROTAING_SENTIVITY;
  pitch = Math.max(minPitch, Math.min(maxPitch, pitch));
  camera.value!.rotation.x = pitch;
}

function onBlur(_e: FocusEvent) {
  document.exitPointerLock();
  isDragging.value = false;
}

export const SpectatorRotation = {
  onPointerDown,
  onBlur,
};
