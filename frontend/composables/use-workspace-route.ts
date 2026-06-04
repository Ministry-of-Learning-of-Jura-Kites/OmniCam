import type { RuntimeConfig } from "nuxt/schema";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export function useWorkspaceRoute(
  projectId: string,
  modelId: string,
  workspaceId: string,
  config: RuntimeConfig,
) {
  const base = getApiBaseUrlWithProtocol("http", config);

  const workspaceBase = concatUrl(
    `projects/${projectId}/models/${modelId}/workspaces`,
    base,
  ).href;
  const workspaceIdUrl = () => concatUrl(workspaceId, workspaceBase).href;
  const me = () => concatUrl("me", workspaceBase).href;
  const merge = () => concatUrl("me/merge", workspaceBase).href;
  const resolve = () => concatUrl("me/resolve", workspaceBase).href;
  const simulation = () => concatUrl("me/simulation", workspaceBase).href;

  return {
    workspaceBase,
    workspaceIdUrl,
    me,
    merge,
    resolve,
    simulation,
  };
}
