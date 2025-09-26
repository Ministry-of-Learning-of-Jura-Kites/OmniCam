<script setup>
import LayoutModelsPage from "~/layouts/ModelsPage.vue";
import Scene3D from "~/components/3d/scene-3d/Scene3D.vue";
import { useRoute } from "vue-router";

const route = useRoute();
const projectId = route.params.projectId;
const modelId = route.params.modelId;
</script>

<template>
  <LayoutModelsPage :workspace="null">
    <Scene3D :project-id="projectId" :model-id="modelId" />
  </LayoutModelsPage>
</template>
