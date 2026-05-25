import type { RuntimeConfig } from "nuxt/schema";
import { MODEL_INFO_KEY } from "~/constants/state-keys";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

function getWorkspaceMeUrl(
  projectId: string,
  modelId: string,
  config: RuntimeConfig,
) {
  const base = getApiBaseUrlWithProtocol("http", config);

  return concatUrl(
    `projects/${projectId}/models/${modelId}/workspaces/me`,
    base,
  ).href;
}

export function useWorkspaceApi(
  projectId: string,
  modelId: string,
  runtimeConfig: RuntimeConfig,
) {
  async function postMerge() {
    const baseUrl = getWorkspaceMeUrl(projectId, modelId, runtimeConfig);
    const resp = await fetch(concatUrl("merge", baseUrl).href, {
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
    try {
      await $fetch<{ error?: string }>(concatUrl("resolve", baseUrl).href, {
        method: "POST",
        credentials: "include",
        body: { merged: results },
      });
    } catch (err) {
      error.value = err as Error;
      console.error("postResolve failed:", err);
    }
    return { error };
  }

  return { postMerge, postCreateMe, deleteWorkspaceMe, postResolve };
}
