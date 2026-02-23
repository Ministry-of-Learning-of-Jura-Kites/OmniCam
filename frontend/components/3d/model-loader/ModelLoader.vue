<!-- eslint-disable @typescript-eslint/no-explicit-any -->
<script setup lang="ts">
import { useGLTF } from "@tresjs/cientos";
import { useFisheye } from "../scene-3d/use-fisheye";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const props = defineProps<{
  path?: string;
  position?: [number, number, number];
}>();

console.log("path", props.path);
const blobUrl = ref<string | null>();
const sceneStates = inject(SCENE_STATES_KEY)!;
let state = null;
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
  const gltf = useGLTF(blobUrl.value);
  state = gltf.state;
} catch (err) {
  console.error("[Fail] load model fail:", err);
}

const { injectFisheye } = useFisheye(sceneStates);

watch(state!, (state) => {
  if (state == null) {
    return;
  }
  state!.scene.traverse((child: any) => {
    if (child.isMesh) {
      child.frustumCulled = false;

      const apply = (mat: any) => {
        mat.onBeforeCompile = injectFisheye;
        // CRITICAL: This forces Three.js to re-read the onBeforeCompile hook
        mat.needsUpdate = true;

        // Optional: Ensure the uniform is unique if you want different strengths
        mat.customProgramCacheKey = () => "fisheye_v1";
      };

      if (Array.isArray(child.material)) {
        child.material.forEach(apply);
      } else {
        apply(child.material);
      }
    }
  });
});
</script>

<template>
  <primitive v-if="state?.scene" :object="state.scene" />

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
