import type { ModelWithCamsResp } from "~/components/3d/scene-states-provider/create-scene-states";
import { MODEL_INFO_KEY } from "~/constants/state-keys";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export function get3dModelPathClient(
  projectId: string,
  modelId: string,
  fileExtension: string,
) {
  const config = useRuntimeConfig();
  const baseApiWithProtocol = getApiBaseUrlWithProtocol("http", config);

  return new URL(
    `assets/projects/${projectId}/models/${modelId}/file/${fileExtension.slice(1)}`,
    baseApiWithProtocol,
  );
}

export function useFetchModel(
  projectId: string,
  modelId: string,
  workspace: string | null,
) {
  const workspaceSuffix = computed(() =>
    workspace ? `/workspaces/${workspace}` : "",
  );
  const runtimeConfig = useRuntimeConfig();
  const nuxtApp = useNuxtApp();
  const key = `${MODEL_INFO_KEY}-${modelId}`;

  const {
    data: modelWithCamsResp,
    error,
    execute,
    status,
  } = useAsyncData<ModelWithCamsResp>(
    key,
    async () => {
      const currentData = nuxtApp.payload.data[key];
      const fields = getRequiredFields(workspace, currentData);

      const headers = useRequestHeaders(["cookie"]);
      const apiBase = getApiBaseUrlWithProtocol("http", runtimeConfig, true);
      const url = new URL(
        `projects/${projectId}/models/${modelId}${workspaceSuffix.value}`,
        apiBase,
      );

      for (const field of fields) {
        url.searchParams.append("fields", field);
      }
      url.searchParams.append("t", Date.now().toString());

      return await $fetch<ModelWithCamsResp>(url.href, {
        headers,
        credentials: "include",
      });
    },
    { immediate: true },
  );

  function getRequiredFields(
    ws: string | null,
    currentData?: ModelWithCamsResp,
  ) {
    if (!currentData)
      return ["cameras", "workspace_exists", "target_area_trapezoids"];
    if (ws == null && currentData.data.workspaceExists == undefined) {
      return ["cameras", "workspace_exists", "target_area_trapezoids"];
    }
    return ["cameras", "target_area_trapezoids"];
  }

  async function fetch() {
    const needsFetch =
      !modelWithCamsResp.value ||
      (workspace == null &&
        modelWithCamsResp.value.data.workspaceExists == undefined) ||
      (workspace == "me" &&
        modelWithCamsResp.value.data.workspaceExists != undefined);

    if (needsFetch) {
      await execute();
    }
  }

  return {
    modelWithCamsResp,
    fetch,
    error,
    pending: computed(() => status.value === "pending"),
  };
}
