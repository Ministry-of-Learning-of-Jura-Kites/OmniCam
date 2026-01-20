<script setup lang="ts">
import { LineBasicMaterial, MeshBasicMaterial, Color, DoubleSide } from "three";
import { useFrustumGeometries } from "~/composables/useFrustumGeometries";
import type { ColorRGBA } from "~/messages/protobufs/autosave_event";

const props = withDefaults(
  defineProps<{
    id: string;
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

const { setFrustumGeometry, removeFrustumGeometry } = useFrustumGeometries();

const lineMaterial = new LineBasicMaterial({ color: 0x000000 });
const meshMaterial = new MeshBasicMaterial({
  color: new Color(
    props.color?.r ?? 0.8,
    props.color?.g ?? 0.8,
    props.color?.b ?? 0.8,
  ),
  transparent: true,
  opacity: props.color?.a ?? 1,
  side: DoubleSide,
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

const frustumGeometries = computed(() => {
  return setFrustumGeometry(props.id, props.fov, props.aspect, props.length);
});

onUnmounted(() => {
  removeFrustumGeometry(props.id);
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
