<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import { Euler } from "three";
import { camera } from "./camera";
import { SpectatorPosition } from "./spectator-position";
import { SpectatorRotation } from "./spectator-rotation";

defineProps<{
  modelId?: string | null;
  placeholderText?: string | null;
}>();

onMounted(() => {
  SpectatorPosition.setup();
});
</script>

<template>
  <ClientOnly>
    <div class="w-full h-full bg-background">
      <TresCanvas
        clear-color="#0E0C29"
        tabindex="0"
        @pointerup="SpectatorRotation.onPointerUp"
        @pointerdown="SpectatorRotation.onPointerDown"
        @pointermove="SpectatorRotation.onPointerMove"
        @keydown="SpectatorPosition.onKeyDown"
        @keyup="SpectatorPosition.onKeyUp"
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
