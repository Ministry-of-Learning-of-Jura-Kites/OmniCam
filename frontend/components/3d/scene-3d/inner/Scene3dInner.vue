<script setup lang="ts">
import { useTresContext } from "@tresjs/core";
import { SCENE_STATES_KEY } from "~/components/3d/scene-states-provider/create-scene-states";

const sceneStates = inject(SCENE_STATES_KEY);
if (!sceneStates) throw new Error("SceneStates is missing!");

onMounted(() => {
  const context = useTresContext();
  console.log("Tres context ready:", context.scene.value);
  sceneStates.tresContext.value = context;
});
</script>

<!-- eslint-disable vue/valid-template-root -->
<template></template>
