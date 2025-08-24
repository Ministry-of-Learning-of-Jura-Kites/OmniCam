<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import type { PerspectiveCamera } from "three";
import { Euler } from "three";

defineProps<{
  modelId?: string | null;
  placeholderText?: string | null;
}>();

const camera: Ref<PerspectiveCamera | null> = ref(null);

let lastX = 0;
let lastY = 0;
const isDragging = ref(false);

function onPointerDown(e: PointerEvent) {
  isDragging.value = true;
  lastX = e.clientX;
  lastY = e.clientY;
}

function onPointerUp(_e: PointerEvent) {
  isDragging.value = false;
}

const maxPitch = Math.PI / 2 - 0.01;
const minPitch = -Math.PI / 2 + 0.01;

let yaw = 0;
let pitch = 0;

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
</script>

<template>
  <ClientOnly>
    <div class="w-full h-full bg-background">
      <TresCanvas
        clear-color="#0E0C29"
        @pointerup="onPointerUp"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
      >
        <!-- Camera -->
        <TresPerspectiveCamera
          ref="camera"
          :position="[4, 4, 4]"
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

        <!-- Camera controls from tresjs/cientos -->
        <!-- <OrbitControls
          :enable-pan="true"
          :enable-zoom="true"
          :enable-rotate="true"
          :min-distance="2"
          :max-distance="50"
        /> -->
      </TresCanvas>
    </div>
  </ClientOnly>
</template>
