<!-- eslint-disable vue/valid-template-root -->

<script setup lang="ts">
import type { WebGLRenderer } from "three";

import type { WatchHandle } from "vue";

import { SCENE_STATES_KEY } from "~/constants/state-keys";

import { createCubeDistortionRenderer } from "~/composables/use-cube-distortion";

const { logger } = useLogger("CubeDistortion");

const sceneStates = inject(SCENE_STATES_KEY);

if (!sceneStates) {
  throw new Error("Expect to be called within scene states provider");
}

let stopWatch: WatchHandle | null = null;
let renderCallback: { off: () => void } | null = null;
let distortion: ReturnType<typeof createCubeDistortionRenderer> | null = null;

onMounted(() => {
  stopWatch = watch(
    () => [sceneStates.value?.tresContext?.value] as const,

    ([tresContext]) => {
      renderCallback?.off();
      renderCallback = null;
      distortion?.dispose();
      distortion = null;

      if (!tresContext) {
        return;
      }

      logger.info("Setting up cube distortion");
      const renderer = tresContext.renderer.instance as WebGLRenderer;
      distortion = createCubeDistortionRenderer(renderer, tresContext.scene);
      sceneStates.value!.cubeCamera.value = distortion.cubeCamera;

      renderCallback = tresContext.renderer.onRender(() => {
        const isDistorting =
          sceneStates.value!.currentDistEnabled.value &&
          sceneStates.value!.transformingInfo.value == undefined;

        if (!isDistorting || !distortion) {
          return;
        }

        const camera = tresContext.camera.activeCamera;

        if (!camera) {
          return;
        }

        renderer.clear();

        distortion.render({
          position: camera.position,
          rotation: camera.rotation,
          fov: sceneStates.value!.currentFov.value,
          aspectRatio: sceneStates.value!.aspectRatio.value,
          width: renderer.domElement.clientWidth,
          height: renderer.domElement.clientHeight,
          isFisheye: sceneStates.value!.currentIsFisheye.value,
          activeCamera: camera,
        });
      });
    },

    { immediate: true },
  );
});

onUnmounted(() => {
  stopWatch?.();

  renderCallback?.off();

  distortion?.dispose();
});
</script>

<template>
  <slot />
</template>
