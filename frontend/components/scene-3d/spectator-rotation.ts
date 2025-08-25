import { camera } from "./camera";

let lastX = 0;
let lastY = 0;
const isDragging = ref(false);

const maxPitch = Math.PI / 2 - 0.01;
const minPitch = -Math.PI / 2 + 0.01;

let yaw = 0;
let pitch = 0;

function onPointerDown(e: PointerEvent) {
  isDragging.value = true;
  lastX = e.clientX;
  lastY = e.clientY;
}

function onPointerUp(_e: PointerEvent) {
  isDragging.value = false;
}

function onPointerMove(e: PointerEvent) {
  if (!isDragging.value) return;
  const deltaX = e.clientX - lastX;
  yaw -= deltaX * 0.01;
  camera.value!.rotation.y = yaw;

  const deltaY = e.clientY - lastY;
  pitch -= deltaY * 0.01;
  pitch = Math.max(minPitch, Math.min(maxPitch, pitch));
  camera.value!.rotation.x = pitch;

  lastX = e.clientX;
  lastY = e.clientY;
}

export const SpectatorRotation = {
  onPointerDown,
  onPointerUp,
  onPointerMove,
};
