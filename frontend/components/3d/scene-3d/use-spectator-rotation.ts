import { SPECTATOR_ROTAING_SENTIVITY } from "~/constants";
import type { SceneStates } from "./use-scene-state";

const maxPitch = Math.PI / 2 - 0.01;
const minPitch = -Math.PI / 2 + 0.01;

export function useSpectatorRotation(sceneStates: SceneStates) {
  const isDragging = ref(false);

  async function onPointerDown(_e: PointerEvent) {
    isDragging.value = true;

    document.addEventListener("pointermove", onPointerMove);
    document.addEventListener("pointerup", onPointerUp);
    await sceneStates.tresCanvasParent.value?.requestPointerLock();
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
    if (
      sceneStates.tresCanvasParent?.value == undefined ||
      sceneStates.isDraggingObject.value
    ) {
      return;
    }

    if (!isDragging.value || sceneStates.currentCameraRotation.value == null)
      return;

    const deltaX = e.movementX;
    let yaw =
      sceneStates.currentCameraRotation.value.y -
      deltaX * SPECTATOR_ROTAING_SENTIVITY;
    yaw = normalizeAngle(yaw);
    sceneStates.currentCameraRotation.value.y = yaw;

    const deltaY = e.movementY;
    let pitch =
      sceneStates.currentCameraRotation.value.x -
      deltaY * SPECTATOR_ROTAING_SENTIVITY;
    pitch = Math.max(minPitch, Math.min(maxPitch, pitch));
    sceneStates.currentCameraRotation.value.x = pitch;
  }

  function onBlur(_e: FocusEvent) {
    document.exitPointerLock();
    isDragging.value = false;
  }

  return {
    onPointerDown,
    onBlur,
  };
}

export type SpectatorRotation = ReturnType<typeof useSpectatorRotation>;
