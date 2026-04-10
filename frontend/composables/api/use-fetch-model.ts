import type { RuntimeConfig } from "nuxt/schema";
import type { ModelWithCamsResp } from "~/components/3d/scene-states-provider/create-scene-states";
import { MODEL_INFO_KEY } from "~/constants/state-keys";
import { getApiBaseUrlWithProtocol } from "~/utils/url";
import type { FetchError } from "ofetch";

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
  workspace: string,
  runtimeConfig: RuntimeConfig,
) {
  const modelWithCamsResp = useState<ModelWithCamsResp | undefined>(
    `${MODEL_INFO_KEY}-${modelId}`,
  );

  const workspaceSuffix = workspace == null ? "" : `/workspaces/${workspace}`;

  const error = useState<FetchError<unknown> | null>(
    "fetch-model-error",
    () => null,
  );

  async function fetchAndCombine(fields: string[]) {
    const paramsObj = {
      fields: fields,
      t: Date.now(),
    };
    const params = objectToQueryParams(paramsObj);
    // Support credentials for both server-side and client-side fetching
    const headers = useRequestHeaders(["cookie"]);
    const apiBaseWithProtocol = getApiBaseUrlWithProtocol(
      "http",
      runtimeConfig,
      true,
    );

    const { data, error: actualError } = await useFetch<ModelWithCamsResp>(
      new URL(
        `projects/${projectId}/models/${modelId}${workspaceSuffix}?${params.toString()}`,
        apiBaseWithProtocol,
      ).href,
      {
        headers: headers,
        credentials: "include",
      },
    );

    modelWithCamsResp.value = data.value;
    error.value = actualError.value as FetchError<unknown>;
  }
  async function fetch() {
    if (modelWithCamsResp.value == undefined) {
      await fetchAndCombine([
        "cameras",
        "workspace_exists",
        "target_area_trapezoids",
      ]);
    } else {
      // If exit from workspace into model
      if (
        workspace == null &&
        modelWithCamsResp.value.data.workspaceExists == undefined
      ) {
        await fetchAndCombine([
          "cameras",
          "workspace_exists",
          "target_area_trapezoids",
        ]);
      }

      // If open workspace from model page
      else if (
        workspace == "me" &&
        modelWithCamsResp.value.data.workspaceExists != undefined
      ) {
        await fetchAndCombine(["cameras", "target_area_trapezoids"]);
      }
    }
  }
  return { modelWithCamsResp, fetch, error };
}
