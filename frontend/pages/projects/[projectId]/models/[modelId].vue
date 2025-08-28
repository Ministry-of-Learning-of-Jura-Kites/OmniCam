<script setup>
import LayoutModelsPage from "@/layouts/modelsPage.vue";
import Scene3D from "@/components/Scene3D.vue";
import { useRoute } from "vue-router";

const route = useRoute();
const modelId = route.params.modelId;
</script>

<template>
  <LayoutModelsPage>
    <Scene3D :model-id="modelId" />
  </LayoutModelsPage>
</template>
