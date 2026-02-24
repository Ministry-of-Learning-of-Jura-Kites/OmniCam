<script setup lang="ts">
import { GLTFModel } from "@tresjs/cientos";
import { CALIBRATION_SCALE, CALIBRATION_HEIGHT } from "~/constants/state-keys";

const props = withDefaults(
  defineProps<{
    path: string;
    position?: [number, number, number];
    modelScale?: number;
  }>(),
  {
    modelScale: 1,
    position: () => [0, 0, 0],
  },
);

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

const scaleFactor = inject(CALIBRATION_SCALE, ref(1));
const modelHeight = inject(CALIBRATION_HEIGHT, ref(0));

const finalScale = computed(() => {
  return props.modelScale * scaleFactor.value;
});
</script>

<template>
  <TresGroup
    :position="[
      props.position?.[0] ?? 0,
      (props.position?.[1] ?? 0) + modelHeight,
      props.position?.[2] ?? 0,
    ]"
    :scale="[finalScale, finalScale, finalScale]"
  >
    <GLTFModel v-if="blobUrl != null" :path="blobUrl" />
  </TresGroup>

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
