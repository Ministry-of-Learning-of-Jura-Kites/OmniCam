<script setup lang="ts">
import type { SceneStates } from "~/types/scene-states";
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

if (sceneStates.error != null) {
  if (
    sceneStates.error &&
    (sceneStates.error as { action: string }).action == "not-found"
  ) {
    throw createError({ statusCode: 404, statusMessage: "Not Found" });
  }
} else {
  const sceneStatesWithHelper = createSceneStatesWithHelper(
    sceneStates as SceneStates,
  );

  provide(SCENE_STATES_KEY, sceneStatesWithHelper);
}
</script>

<template>
  <slot />
</template>
