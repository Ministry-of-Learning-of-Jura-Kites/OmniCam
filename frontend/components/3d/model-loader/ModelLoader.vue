<script setup lang="ts">
import { GLTFModel } from "@tresjs/cientos";

const props = defineProps<{
  path?: string;
  position?: [number, number, number];
}>();

console.log("path", props.path);
const blobUrl = ref<string | null>();
try {
  const response = await useFetch(props.path ?? "", {
    method: "GET",
    credentials: "include", // <-- Important! sends cookies/session
    responseType: "blob",
  });

  if (response.error.value != undefined) {
    throw new Error(`HTTP ${response.status.value} ${response.status.value}`);
  }

  const blob = (await response.data.value) as Blob;
  if (!blob || blob.size === 0) {
    throw new Error("Blob is empty or invalid");
  }
  blobUrl.value = URL.createObjectURL(blob);
} catch (err) {
  console.error("[Fail] load model fail:", err);
}
</script>

<template>
  <GLTFModel v-if="blobUrl != null" :path="blobUrl" />

  <!-- Block Placeholder  -->
  <TresMesh v-if="blobUrl == null" :position="props.position ?? [0, 0, 0]">
    <TresBoxGeometry />
    <TresMeshStandardMaterial
      :color="'#4a90e2'"
      :metalness="0.3"
      :roughness="0.4"
    />
  </TresMesh>
</template>
