<script setup lang="ts">
import type { SceneStates } from "~/types/scene-states";
import {
  createBaseSceneStates,
  createSceneStatesWithHelper,
  type ModelWithCamsResp,
} from "./create-scene-states";
import { useWebSocket, type UseWebSocketReturn } from "@vueuse/core";
import {
  MODEL_INFO_KEY,
  SCENE_STATES_KEY,
  CALIBRATION_SCALE,
  CALIBRATION_HEIGHT,
} from "~/constants/state-keys";

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

const modelWithCamsResp = useState<ModelWithCamsResp | undefined>(
  `${MODEL_INFO_KEY}-${props.modelId}`,
);

const error = ref<unknown | undefined>(undefined);

async function fetchAndCombine(fields: string[]) {
  const paramsObj = {
    fields: fields,
    t: Date.now(),
  };
  const params = objectToQueryParams(paramsObj);
  // Support credentials for both server-side and client-side fetching
  const headers = useRequestHeaders(["cookie"]);

  const { data: fetchedModelWithCamsResp, error: modelFetchError } =
    await useAsyncData("model_information", () =>
      $fetch<ModelWithCamsResp>(
        `http://${getHostFromRuntime(runtimeConfig, import.meta.client)}/api/v1/projects/${props.projectId}/models/${props.modelId}${workspaceSuffix}?${params.toString()}`,
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
      statusMessage: modelFetchError.value.statusMessage + " " + error.value,
      fatal: true,
    });
  }

  modelWithCamsResp.value = {
    ...modelWithCamsResp.value,
    ...fetchedModelWithCamsResp.value!,
  };
}

if (modelWithCamsResp.value == undefined) {
  await fetchAndCombine(["cameras", "workspace_exists"]);
} else {
  // If exit from workspace into model
  if (
    props.workspace == null &&
    modelWithCamsResp.value.data.workspaceExists == undefined
  ) {
    await fetchAndCombine(["cameras", "workspace_exists"]);
  }

  // If open workspace from model page
  else if (
    props.workspace != null &&
    modelWithCamsResp.value.data.workspaceExists != undefined
  ) {
    await fetchAndCombine(["cameras"]);
  }
}

const calibrationScale = inject<Ref<number>>(CALIBRATION_SCALE)!;
const calibrationHeight = inject<Ref<number>>(CALIBRATION_HEIGHT)!;

if (modelWithCamsResp.value?.data.scaleFactor !== undefined) {
  calibrationScale.value = modelWithCamsResp.value.data.scaleFactor;
}
if (modelWithCamsResp.value?.data.modelHeight !== undefined) {
  calibrationHeight.value = modelWithCamsResp.value.data.modelHeight;
}

let websocket: UseWebSocketReturn<unknown> | undefined = undefined;
if (props.workspace != undefined && import.meta.client) {
  const websocketUrl = `ws://${runtimeConfig.public.externalBackendHost}/api/v1/projects/${props.projectId}/models/${props.modelId}/autosave`;

  websocket = useWebSocket(websocketUrl, {
    autoReconnect: {
      delay: 1000,
      onFailed: () => {
        alert("Failed to connect websocket after multiple retries.");
      },
    },
  });
}

const sceneStates = createBaseSceneStates(
  websocket,
  modelWithCamsResp.value!,
  calibrationScale,
  calibrationHeight,
);

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
