<script setup lang="ts">
import { useGLTF } from "@tresjs/cientos";
import type { TresObject3D } from "@tresjs/core";

const props = defineProps<{
  path?: string;
  position?: [number, number, number];
}>();

let nodes: Record<string, TresObject3D>;

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
    v-for="[key, node] of Object.entries(nodes)"
    :key="key"
    :object="node"
    :position="[
      (props.position?.[0] ?? 0) + (node.position?.x ?? 0),
      (props.position?.[1] ?? 0) + (node.position?.y ?? 0),
      (props.position?.[2] ?? 0) + (node.position?.z ?? 0),
    ]"
  />

  <!-- Block Placeholder  -->
  <TresMesh v-if="nodes == undefined" :position="props.position ?? [0, 0, 0]">
    <TresBoxGeometry />
    <TresMeshStandardMaterial
      :color="'#4a90e2'"
      :metalness="0.3"
      :roughness="0.4"
    />
  </TresMesh>
</template>
