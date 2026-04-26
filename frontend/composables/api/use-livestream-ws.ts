import { type UseWebSocketReturn, useWebSocket } from "@vueuse/core";
import type { RuntimeConfig } from "nuxt/schema";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export function useLivestreamWs(
  modelId: string,
  workspace: string,
  runtimeConfig: RuntimeConfig,
) {
  let livestreamWs: UseWebSocketReturn<unknown> | undefined = undefined;
  if (workspace != undefined && workspace != "me" && import.meta.client) {
    const base = getApiBaseUrlWithProtocol("websocket", runtimeConfig);
    const websocketUrl = concatUrl(
      `models/${modelId}/livestream/${workspace}`,
      base,
    );
    // TODO: Fix, don't hard code api/v1
    livestreamWs = useWebSocket(websocketUrl, {
      autoReconnect: {
        delay: 1000,
        onFailed: () => {
          alert("Failed to connect websocket after multiple retries.");
        },
      },
    });
  }
  return { livestreamWs };
}
