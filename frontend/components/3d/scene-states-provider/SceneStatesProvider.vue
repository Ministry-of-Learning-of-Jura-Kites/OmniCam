<script setup lang="ts">
import type { SceneStates } from "~/types/scene-states";
import {
  createBaseSceneStates,
  createSceneStatesWithHelper,
  SCENE_STATES_KEY,
  type ModelWithCamsResp,
} from "./create-scene-states";
import { useWebSocket, type UseWebSocketReturn } from "@vueuse/core";

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

const workspaceSuffix =
  props.workspace == null ? "" : `/workspaces/${props.workspace}`;

const fields = ["cameras"];

const params = new URLSearchParams();
for (const field of fields) {
  params.append("fields", field);
}
params.append("t", String(Date.now()));

const { data: modelWithCamsResp, error: modelFetchError } =
  await useAsyncData<ModelWithCamsResp>("model_information", () =>
    $fetch(
      `http://${runtimeConfig.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${props.projectId}/models/${props.modelId}${workspaceSuffix}?${params.toString()}`,
    ),
  );

if (modelFetchError.value != undefined) {
  showError({
    statusCode: 404,
    statusMessage: "Not Found",
    fatal: true,
  });
}

const websocketUrl = `ws://${runtimeConfig.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${props.projectId}/models/${props.modelId}/autosave`;

let websocket: UseWebSocketReturn<unknown> | undefined = undefined;
if (props.workspace != undefined) {
  websocket = useWebSocket(websocketUrl, {
    autoReconnect: {
      delay: 1000,
      onFailed: () => {
        alert("Failed to connect websocket after multiple retries.");
      },
    },
  });
}

const sceneStates = createBaseSceneStates(websocket, modelWithCamsResp.value!);

if (sceneStates.error != null) {
  if (
    sceneStates.error &&
    (sceneStates.error as { action: string }).action == "not-found"
  ) {
    showError({
      statusCode: 404,
      statusMessage: "Not Found",
      fatal: true,
    });
  }
} else {
  const sceneStatesWithHelper = createSceneStatesWithHelper(
    sceneStates as SceneStates,
  );

  provide(SCENE_STATES_KEY, sceneStatesWithHelper);
}
</script>

<template>
  <slot v-if="modelFetchError == undefined" />
</template>
