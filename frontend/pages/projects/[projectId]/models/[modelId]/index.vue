<script setup>
import Scene3D from "~/components/3d/scene-3d/Scene3D.vue";
import { useRoute } from "vue-router";

const workspace = null;

definePageMeta({
  layout: "models-page",
  routeInfo: {
    workspace: workspace,
  },
});

const route = useRoute();
const projectId = route.params.projectId;
const modelId = route.params.modelId;
</script>

<template>
  <Scene3D :project-id="projectId" :model-id="modelId" :workspace="workspace" />
</template>
