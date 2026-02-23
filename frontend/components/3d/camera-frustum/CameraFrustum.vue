<script setup lang="ts">
import {
  LineBasicMaterial,
  MeshBasicMaterial,
  Color,
  DoubleSide,
  type BufferGeometry,
} from "three";
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
    color: () => ({ r: 1, g: 0.8, b: 0.2, a: 0.3 }),
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
  opacity: props.color?.a ?? 0.3,
  side: DoubleSide,
  depthWrite: false,
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

const frustumGeometries = ref<{
  mesh: BufferGeometry;
  lines: BufferGeometry;
} | null>(null);

watch(
  [() => props.id, () => props.fov, () => props.aspect, () => props.length],
  ([id, fov, aspect, length]) => {
    // This calls your composable logic
    const result = setFrustumGeometry(id, fov, aspect, length);

    // Update the local ref for the template
    frustumGeometries.value = result;
  },
  { immediate: true },
);

onUnmounted(() => {
  removeFrustumGeometry(props.id);
  lineMaterial.dispose();
  meshMaterial.dispose();
});
</script>

<template>
  <template v-if="frustumGeometries">
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
</template>
