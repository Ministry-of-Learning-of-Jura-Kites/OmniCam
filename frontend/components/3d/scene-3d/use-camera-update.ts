import type { SceneStates } from "./use-scene-state";

export function useCameraUpdate(states: SceneStates) {
  watch(
    states.currentCameraPosition,
    (pos) => {
      if (states.tresContext.value?.camera) {
        states.tresContext.value.camera.position.set(pos.x, pos.y, pos.z);
      }
    },
    { deep: true },
  );

  watch(
    states.currentCameraRotation,
    (pos) => {
      if (states.tresContext.value?.camera) {
        states.tresContext.value.camera.rotation.set(pos.x, pos.y, pos.z);
      }
    },
    { deep: true },
  );
}
