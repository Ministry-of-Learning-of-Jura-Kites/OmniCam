<script setup lang="ts">
import {
  createBaseSceneStates,
  createSceneStatesWithHelper,
  SCENE_STATES_KEY,
} from "./create-scene-states";

const props = defineProps({
  projectId: {
    type: String,
    required: true,
  },
  modelId: {
    type: String,
    required: true,
  },
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = await createBaseSceneStates(
  props.projectId,
  props.modelId,
  useRuntimeConfig(),
  props.workspace,
);

const sceneStatesWithHelper = createSceneStatesWithHelper(sceneStates);

provide(SCENE_STATES_KEY, sceneStatesWithHelper);
</script>

<template>
  <slot />
</template>
