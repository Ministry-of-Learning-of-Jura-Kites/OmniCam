<script setup>
import Scene3D from "~/components/3d/scene-3d/Scene3D.vue";
import { useRoute } from "vue-router";
import { WORKSPACE_KEY } from "~/layouts/workspace-provider";

definePageMeta({
  layout: "models-page",
});

const route = useRoute();
const projectId = route.params.projectId;
const modelId = route.params.modelId;

const workspace = inject(WORKSPACE_KEY);
workspace.value = "me";
</script>

<template>
  <Scene3D :project-id="projectId" :model-id="modelId" workspace="me" />
</template>
