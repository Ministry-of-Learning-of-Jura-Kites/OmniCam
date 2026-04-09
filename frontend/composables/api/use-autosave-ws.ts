import { useWebSocket, type UseWebSocketReturn } from "@vueuse/core";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export function useAutosaveWs(
  projectId: string,
  modelId: string,
  workspace: string,
) {
  const runtimeConfig = useRuntimeConfig();
  let autosaveWs: UseWebSocketReturn<unknown> | undefined = undefined;
  if (workspace == "me" && import.meta.client) {
    const base = getApiBaseUrlWithProtocol("websocket", runtimeConfig);
    const websocketUrl = new URL(
      `projects/${projectId}/models/${modelId}/autosave`,
      base,
    );

    autosaveWs = useWebSocket(websocketUrl, {
      autoReconnect: {
        delay: 1000,
        onFailed: () => {
          alert("Failed to connect websocket after multiple retries.");
        },
      },
    });
  }

  return { autosaveWs };
}
