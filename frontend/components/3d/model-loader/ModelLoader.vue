<script setup lang="ts">
import { useGLTF } from "@tresjs/cientos";

const props = defineProps<{
  path?: string;
  position?: [number, number, number];
}>();

let nodes: any;

try {
  const gltf = await useGLTF(props.path ?? "");
  nodes = gltf.nodes;
  console.log("[Success] load model success:", nodes);
} catch (err) {
  console.error("[Fail] load model fail:", err);
}
</script>

<template>
  <primitive
    v-if="nodes"
    :object="nodes.mesh_0"
    :position="props.position ?? [0, 0, 0]"
  />

  <!-- Block Placeholder  -->
  <TresMesh v-if="!nodes" :position="props.position ?? [0, 0, 0]">
    <TresBoxGeometry />
    <TresMeshStandardMaterial
      :color="'#4a90e2'"
      :metalness="0.3"
      :roughness="0.4"
    />
  </TresMesh>
</template>
