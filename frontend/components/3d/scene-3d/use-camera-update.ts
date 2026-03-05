import type { SceneStatesWithHelper } from "~/types/scene-states";

export function useCameraUpdate(states: SceneStatesWithHelper) {
  watch(
    () => [states.transformingInfo, states.currentCam],
    ([transform, cam]) => {
      const rotation =
        transform?.value == undefined
          ? cam!.value!.rotation
          : transform.value.rotation;
      if (states.tresContext.value?.camera) {
        states.tresContext.value.camera.rotation.set(
          rotation.x,
          rotation.y,
          rotation.z,
        );
      }

      const pos =
        transform?.value == undefined
          ? cam!.value!.position
          : transform.value.position;
      if (states.tresContext.value?.camera) {
        states.tresContext.value.camera.position.set(pos.x, pos.y, pos.z);
      }
    },
    { deep: true },
  );
}
