<script setup lang="ts">
import { useTresContext } from "@tresjs/core";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

const sceneStates = inject(SCENE_STATES_KEY);
if (!sceneStates) throw new Error("SceneStates is missing!");

onMounted(() => {
  const context = useTresContext();
  sceneStates.tresContext.value = context;
});
</script>

<!-- eslint-disable vue/valid-template-root -->
<template></template>
