<script setup lang="ts">
import type { SceneStates, SceneStatesWithHelper } from "~/types/scene-states";
import {
  createBaseSceneStates,
  createSceneStatesWithHelper,
} from "./create-scene-states";
import {
  SCENE_STATES_KEY,
  SCENE_STATES_READY_KEY,
} from "~/constants/state-keys";
import { useFetchModel } from "~/composables/api/use-fetch-model";
import { useAutosaveWs } from "~/composables/api/use-autosave-ws";
import { useLivestreamWs } from "~/composables/api/use-livestream-ws";

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

const runtimeConfig = useRuntimeConfig();

const {
  modelWithCamsResp,
  error,
  fetch: fetchModel,
} = useFetchModel(props.projectId, props.modelId, props.workspace);

const { autosaveWs } = useAutosaveWs(
  props.projectId,
  props.modelId,
  props.workspace,
);

const { livestreamWs } = useLivestreamWs(
  props.modelId,
  props.workspace,
  runtimeConfig,
);

const sceneStates = shallowRef<SceneStatesWithHelper | undefined>(undefined);
const sceneStatesReady = inject(SCENE_STATES_READY_KEY);
provide(SCENE_STATES_KEY, sceneStates);

await fetchModel();

if (error.value != null) {
  showError(error.value);
}

const baseSceneStates = createBaseSceneStates(
  autosaveWs,
  livestreamWs,
  modelWithCamsResp.value!,
);

if (baseSceneStates.error != null) {
  if (
    baseSceneStates.error &&
    (baseSceneStates.error as { action: string }).action == "not-found"
  ) {
    showError({
      statusCode: 404,
      statusMessage: "Not Found",
      fatal: true,
    });
  }
} else {
  sceneStates.value = createSceneStatesWithHelper(
    baseSceneStates as SceneStates,
    props.workspace,
  );
  await nextTick();
  sceneStatesReady!.value = true;
}

if (error.value != null) {
  showError(error.value);
}
</script>

<template>
  <slot v-if="error == undefined" />
</template>
