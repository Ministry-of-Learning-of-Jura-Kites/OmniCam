<script setup lang="ts">
import * as THREE from "three";
import { createFrustumGeometry } from "./create-frustum";
import { boolean } from "zod";

const props = withDefaults(
  defineProps<{
    fov?: number;
    aspect?: number; //(width / height)
    length?: number; //Far Plane
    isHiding?: boolean;
  }>(),
  {
    fov: 60,
    aspect: 16 / 9,
    length: 1e6,
    isHiding: false,
  },
);

const frustumGeometry = computed(() => {
  return createFrustumGeometry(props.fov, props.aspect, props.length);
});
</script>

<template>
  <TresLineSegments
    :geometry="frustumGeometry"
    :material="new THREE.LineBasicMaterial({ color: 0x400c00 })"
    :visible="!isHiding"
  />
  <TresMesh
    :geometry="frustumGeometry"
    :material="new THREE.LineBasicMaterial({ color: 0xff7a5c })"
    :visible="!isHiding"
  />
</template>
