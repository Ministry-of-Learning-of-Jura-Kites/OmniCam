<script setup lang="ts">
import LazyScene3D from "~/components/3d/scene-3d/Scene3D.vue";
import { useRoute } from "vue-router";

const workspace = "me";

definePageMeta({
  layout: "models-page",
  routeInfo: {
    workspace: workspace,
  },
});

const route = useRoute();
const projectId = route.params.projectId as string;
const modelId = route.params.modelId as string;
</script>

<template>
  <LazyScene3D
    :project-id="projectId"
    :model-id="modelId"
    :workspace="workspace"
  />
</template>
