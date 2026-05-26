import { type UseWebSocketReturn, useWebSocket } from "@vueuse/core";
import type { RuntimeConfig } from "nuxt/schema";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export function useLivestreamWs(
  modelId: string,
  workspace: string,
  runtimeConfig: RuntimeConfig,
) {
  let livestreamWs: UseWebSocketReturn<unknown> | undefined = undefined;
  const livestreamError = ref<string | null>("");
  if (workspace != undefined && workspace != "me" && import.meta.client) {
    const base = getApiBaseUrlWithProtocol("websocket", runtimeConfig);
    const websocketUrl = concatUrl(
      `models/${modelId}/livestream/${workspace}`,
      base,
    );
    livestreamWs = useWebSocket(websocketUrl, {
      autoReconnect: {
        delay: 1000,
        onFailed: () => {
          alert("Failed to connect websocket after multiple retries.");
        },
      },
      onDisconnected: (ws, event) => {
        const closeEvent = event as CloseEvent;
        if (closeEvent?.code === 1011) {
          livestreamError.value =
            closeEvent.reason || "Livestream subscription failed";
        }
      },
    });
  }
  return { livestreamWs, livestreamError };
}
