<script setup>
import LazyScene3D from "~/components/3d/scene-3d/Scene3D.vue";
import { useRoute } from "vue-router";
import { SCENE_STATES_READY_KEY } from "~/constants/state-keys";

definePageMeta({
  layout: "models-page",
});

const sceneStatesReady = inject(SCENE_STATES_READY_KEY);

const route = useRoute();
const projectId = route.params.projectId;
const modelId = route.params.modelId;
const workspace = route.params.workspaceId;
</script>

<template>
  <LazyScene3D
    v-if="sceneStatesReady"
    :project-id="projectId"
    :model-id="modelId"
    :workspace="workspace"
  />
</template>
