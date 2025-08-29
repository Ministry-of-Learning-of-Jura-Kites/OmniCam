import type { SceneStatesWithHelper } from "~/types/scene-states";

export function useCameraUpdate(states: SceneStatesWithHelper) {
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
