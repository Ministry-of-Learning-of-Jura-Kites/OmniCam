<script setup lang="ts">
import type { SceneStates } from "~/types/scene-states";
import {
  createBaseSceneStates,
  createSceneStatesWithHelper,
  SCENE_STATES_KEY,
  type ModelWithCamsResp,
} from "./create-scene-states";
import { useWebSocket, type UseWebSocketReturn } from "@vueuse/core";
import { MODEL_INFO_KEY } from "~/constants/state-keys";

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

const fields = ["cameras", "workspace_exists"];

const params = new URLSearchParams();
for (const field of fields) {
  params.append("fields", field);
}
params.append("t", String(Date.now()));

const modelWithCamsResp = useState<ModelWithCamsResp | undefined>(
  MODEL_INFO_KEY,
);
const error = ref<unknown | undefined>(undefined);

if (modelWithCamsResp.value == undefined && import.meta.server) {
  // Support credentials for both server-side and client-side fetching
  const headers = useRequestHeaders(["cookie"]);

  const { data: fetchedModelWithCamsResp, error: modelFetchError } =
    await useAsyncData("model_information", () =>
      $fetch<ModelWithCamsResp>(
        `http://${runtimeConfig.internalBackendHost}/api/v1/projects/${props.projectId}/models/${props.modelId}${workspaceSuffix}?${params.toString()}`,
        {
          headers: headers,
          credentials: "include",
        },
      ),
    );

  if (modelFetchError.value != undefined) {
    error.value = modelFetchError.value;
    showError({
      statusCode: modelFetchError.value.statusCode,
      statusMessage: modelFetchError.value.statusMessage,
      fatal: true,
    });
  }

  modelWithCamsResp.value = fetchedModelWithCamsResp.value;
}

const websocketUrl = `ws://${runtimeConfig.public.externalBackendHost}/api/v1/projects/${props.projectId}/models/${props.modelId}/autosave`;

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
    props.workspace,
  );

  provide(SCENE_STATES_KEY, sceneStatesWithHelper);
}
</script>

<template>
  <slot v-if="error == undefined" />
</template>
