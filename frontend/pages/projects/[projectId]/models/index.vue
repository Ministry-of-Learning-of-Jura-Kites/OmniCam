<template>
  <LayoutModelsPage>
    <Scene3D :modelId="null" placeholderText="Select a model to display" />
  </LayoutModelsPage>
</template>

<script setup>
import LayoutModelsPage from "@/layouts/modelsPage.vue";
import Scene3D from "@/components/Scene3D.vue";
</script>
