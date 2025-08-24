<template>
  <ClientOnly>
    <div class="w-full h-full bg-background">
      <TresCanvas clear-color="#0E0C29">
        <!-- Camera -->
        <TresPerspectiveCamera
          :position="[4, 4, 4]"
          :look-at="[0, 1, 0]"
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
        <OrbitControls
          :enable-pan="true"
          :enable-zoom="true"
          :enable-rotate="true"
          :min-distance="2"
          :max-distance="50"
        />
      </TresCanvas>
    </div>
  </ClientOnly>
</template>

<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { OrbitControls, Grid, Environment } from "@tresjs/cientos";
import { ref } from "vue";

defineProps<{
  modelId?: string | null;
  placeholderText?: string | null;
}>();
</script>
