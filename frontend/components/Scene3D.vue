<template>
  <ClientOnly>
    <div class="w-full h-full bg-background">
      <TresCanvas clear-color="#82DBC5">
        <!-- Camera -->
        <TresPerspectiveCamera
          :position="[4, 4, 4]"
          :look-at="[0, 1, 0]"
          :fov="75"
        />
        <!-- 3D Objects -->
        <TresMesh :position="[0, 0, 0]">
          <TresBoxGeometry />
          <TresMeshStandardMaterial />
        </TresMesh>

        <!-- Grid component from tresjs/cientos, similar to @react-three/drei's Grid -->
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

<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { OrbitControls, Grid, Environment } from "@tresjs/cientos";
import { ref } from "vue";

defineProps<{
  modelId?: string | null;
  placeholderText?: string | null;
}>();
</script>
