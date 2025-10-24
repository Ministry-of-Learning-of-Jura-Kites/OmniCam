<script setup lang="ts">
import * as THREE from "three";
import { createFrustumGeometry } from "./create-frustum";
import type { ColorRGBA } from "~/messages/protobufs/autosave_event";

const props = withDefaults(
  defineProps<{
    fov?: number;
    aspect?: number; //(width / height)
    length?: number; //Far Plane
    color?: ColorRGBA;
    isHiding?: boolean;
  }>(),
  {
    fov: 60,
    aspect: 16 / 9,
    length: 1e6,
    color: () => ({ r: 1, g: 0.8, b: 0.2, a: 0.5 }),
    isHiding: false,
  },
);
const lineMaterial = new THREE.LineBasicMaterial({ color: 0x000000 });
const meshMaterial = new THREE.MeshBasicMaterial({
  color: new THREE.Color(
    props.color?.r ?? 0.8,
    props.color?.g ?? 0.8,
    props.color?.b ?? 0.8,
  ),
  transparent: true,
  opacity: props.color?.a ?? 1,
  side: THREE.DoubleSide,
});

watch(
  () => props.color,
  (color) => {
    if (!color) return;
    meshMaterial.color.setRGB(color.r, color.g, color.b);
    meshMaterial.opacity = color.a ?? 1;
  },
  { deep: true, immediate: true },
);

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
