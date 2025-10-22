<script setup lang="ts">
import * as THREE from "three";
import { createFrustumGeometry } from "./create-frustum";

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
const lineMaterial = new THREE.LineBasicMaterial({ color: 0x400c00 });
const meshMaterial = new THREE.MeshBasicMaterial({
  color: 0xff7a5c,
  transparent: false,
  opacity: 1,
  side: THREE.DoubleSide,
});

let oldGeometries: ReturnType<typeof createFrustumGeometry> | null = null;

const frustumGeometries = computed(() => {
  const pair = createFrustumGeometry(props.fov, props.aspect, props.length);

  // Dispose of old geometry if it exists
  if (oldGeometries) {
    oldGeometries.mesh.dispose();
    oldGeometries.lines.dispose();
  }
  oldGeometries = pair;

  return pair;
});

onUnmounted(() => {
  if (oldGeometries) {
    oldGeometries.mesh.dispose();
    oldGeometries.lines.dispose();
  }
  lineMaterial.dispose();
  meshMaterial.dispose();
});
</script>

<template>
  <TresLineSegments
    :geometry="frustumGeometries.lines"
    :material="lineMaterial"
    :visible="!isHiding"
  />
  <TresMesh
    :geometry="frustumGeometries.mesh"
    :material="meshMaterial"
    :visible="!isHiding"
  />
</template>
