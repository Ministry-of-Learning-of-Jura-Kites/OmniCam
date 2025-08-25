<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import { Euler } from "three";
import { camera, tresCanvasParent, cameraPosition } from "./refs";
import { SpectatorPosition } from "./spectator-position";
import { SpectatorRotation } from "./spectator-rotation";
import AdjustableInput from "../adjustable-input/AdjustableInput.vue";

defineProps<{
  modelId?: string | null;
  placeholderText?: string | null;
}>();

onMounted(() => {
  SpectatorPosition.setup();
});

watch(
  cameraPosition,
  (pos) => {
    if (camera.value) {
      camera.value.position.set(pos.x, pos.y, pos.z);
    }
  },
  { deep: true },
);
</script>

<template>
  <ClientOnly>
    <div class="w-full h-full bg-background relative" ref="tresCanvasParent">
      <div
        id="camera-props"
        class="absolute top-0 right-0 z-10 text-white flex flex-col p-2"
      >
        <p>Camera</p>
        <div class="flex">
          <p>x:</p>
          <AdjustableInput
            v-model="cameraPosition.x"
            class="text-right pl-2"
          ></AdjustableInput>
        </div>
        <div class="flex">
          <p>y:</p>
          <AdjustableInput
            v-model="cameraPosition.y"
            class="text-right pl-2"
          ></AdjustableInput>
        </div>
        <div class="flex">
          <p>z:</p>
          <AdjustableInput
            v-model="cameraPosition.z"
            class="text-right pl-2"
          ></AdjustableInput>
        </div>
      </div>
      <TresCanvas
        id="canvas"
        ref="canvas"
        resize-event="parent"
        clear-color="#0E0C29"
        tabindex="0"
        @pointerdown="SpectatorRotation.onPointerDown"
        @keydown="SpectatorPosition.onKeyDown"
        @keyup="SpectatorPosition.onKeyUp"
        @blur="
          (event: any) => {
            SpectatorRotation.onBlur(event);
            SpectatorPosition.onBlur(event);
          }
        "
      >
        <!-- Camera -->
        <TresPerspectiveCamera
          ref="camera"
          :position="cameraPosition"
          :rotation="new Euler(0, 0, 0, 'YXZ')"
          :fov="75"
        />

        <!-- Environment and lighting, from the tresjs/cientos library -->
        <Environment preset="studio" />
        <TresAmbientLight :intensity="0.4" />
        <TresDirectionalLight :position="[10, 10, 5]" :intensity="1" />

        <!-- 3D Objects -->
        <TresMesh ref="meshRef" :position="[0, 0.5, 0]">
          <TresBoxGeometry />
          <TresMeshStandardMaterial
            :color="'#4a90e2'"
            :metalness="0.3"
            :roughness="0.4"
          />
        </TresMesh>

        <!-- Grid -->
        <Grid
          :args="[20, 20]"
          :cell-size="1"
          :cell-thickness="0.5"
          :cell-color="'#4a90e2'"
          :section-size="5"
          :section-thickness="1"
          :section-color="'#ffffff'"
          :fade-distance="50"
          :fade-strength="1"
          infinite-grid
        />
      </TresCanvas>
    </div>
  </ClientOnly>
</template>

<style>
#canvas {
  height: 100%;
  width: 100%;
  min-height: 0;
}
#camera-props {
  text-shadow:
    -1px -1px 0 black /* Top-left shadow */,
    1px -1px 0 black /* Top-right shadow */,
    -1px 1px 0 black /* Bottom-left shadow */,
    1px 1px 0 black /* Bottom-right shadow */;
}
</style>
