<script setup>
import LayoutModelsPage from "~/layouts/ModelsPage.vue";
import Scene3D from "~/components/3d/scene-3d/Scene3D.vue";
import { useRoute } from "vue-router";

const route = useRoute();
const modelId = route.params.modelId;
</script>

<template>
  <LayoutModelsPage>
    <Scene3D :model-id="modelId" />
  </LayoutModelsPage>
</template>
