<script setup lang="ts">
import {
  createBaseSceneStates,
  createSceneStatesWithHelper,
  SCENE_STATES_KEY,
} from "./create-scene-states";

const sceneStates = createBaseSceneStates();

const sceneStatesWithHelper = createSceneStatesWithHelper(sceneStates);

provide(SCENE_STATES_KEY, sceneStatesWithHelper);
</script>

<template>
  <slot />
</template>
