<script setup lang="ts">
import * as THREE from "three";
import MovableArrow from "../movable-arrow/MovableArrow.vue";

const props = defineProps({
  name: {
    type: String,
    default: "Untitled",
  },
  position: {
    type: THREE.Vector3,
    required: true,
  },
  rotation: {
    type: THREE.Euler,
    required: true,
  },
});

const cameraMesh = ref<THREE.Mesh | null>(null);
</script>

<template>
  <TresMesh ref="cameraMesh" :position="props.position">
    <TresMesh :rotation="props.rotation">
      <TresBoxGeometry :args="[0.5, 0.5, 0.5]" />
      <TresMeshBasicMaterial color="white" />
    </TresMesh>
    <!-- <primitive :object="arrow" /> -->
    <MovableArrow
      v-if="cameraMesh != null"
      direction="x"
      :camera-mesh="cameraMesh"
      color="green"
    />
    <MovableArrow
      v-if="cameraMesh != null"
      direction="y"
      :camera-mesh="cameraMesh"
      color="red"
    />
    <MovableArrow
      v-if="cameraMesh != null"
      direction="z"
      :camera-mesh="cameraMesh"
      color="blue"
    />
  </TresMesh>
</template>
