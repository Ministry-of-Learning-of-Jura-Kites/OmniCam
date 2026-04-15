import type { RuntimeConfig } from "nuxt/schema";
import { MODEL_INFO_KEY } from "~/constants/state-keys";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

function getWorkspaceMeUrl(
  projectId: string,
  modelId: string,
  config: RuntimeConfig,
) {
  const base = getApiBaseUrlWithProtocol("http", config, true);

  return new URL(`projects/${projectId}/models/${modelId}/workspaces/me`, base)
    .href;
}

export function useWorkspaceApi(
  projectId: string,
  modelId: string,
  runtimeConfig: RuntimeConfig,
) {
  async function postMerge() {
    const baseUrl = getWorkspaceMeUrl(projectId, modelId, runtimeConfig);
    const resp = await fetch(new URL("merge", baseUrl).href, {
      method: "POST",
      credentials: "include",
    });
    return { resp };
  }

  async function postCreateMe() {
    const data = await $fetch(
      getWorkspaceMeUrl(projectId, modelId, runtimeConfig),
      {
        method: "POST",
        credentials: "include",
      },
    );
    useState(`${MODEL_INFO_KEY}-${modelId}`, () => data);
  }

  async function deleteWorkspaceMe() {
    await $fetch(getWorkspaceMeUrl(projectId, modelId, runtimeConfig), {
      method: "DELETE",
      credentials: "include",
    });
    useState(`${MODEL_INFO_KEY}-${modelId}`, () => undefined);
  }

  async function postResolve(results: Record<string, Record<string, unknown>>) {
    const error = ref<Error | undefined>();
    const baseUrl = getWorkspaceMeUrl(projectId, modelId, runtimeConfig);
    await $fetch<{ error?: string }>(new URL("resolve", baseUrl).href, {
      method: "POST",
      credentials: "include",
      body: { merged: results },
      onResponseError: (ctx) => {
        error.value = ctx.error;
      },
    });
    return { error };
  }

  return { postMerge, postCreateMe, deleteWorkspaceMe, postResolve };
}
