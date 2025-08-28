<script setup>
import LayoutModelsPage from "@/layouts/modelsPage.vue";
import Scene3D from "@/components/3d/Scene3D.vue";
</script>

<template>
  <LayoutModelsPage>
    <Scene3D :model-id="null" placeholder-text="Select a model to display" />
  </LayoutModelsPage>
</template>
