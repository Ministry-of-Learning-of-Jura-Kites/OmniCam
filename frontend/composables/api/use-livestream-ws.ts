import { type UseWebSocketReturn, useWebSocket } from "@vueuse/core";
import type { RuntimeConfig } from "nuxt/schema";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export function useLivestreamWs(
  modelId: string,
  workspace: string,
  runtimeConfig: RuntimeConfig,
) {
  let livestreamWs: UseWebSocketReturn<unknown> | undefined = undefined;
  const livestreamAlert = ref<{
    title: string;
    message: string;
    type: "success" | "warning" | "error";
  } | null>(null);
  if (workspace != undefined && workspace != "me" && import.meta.client) {
    const base = getApiBaseUrlWithProtocol("websocket", runtimeConfig);
    const websocketUrl = concatUrl(
      `models/${modelId}/livestream/${workspace}`,
      base,
    );

    let wasConnected = false;
    let isDisconnectedAlertVisible = false;

    livestreamWs = useWebSocket(websocketUrl, {
      autoReconnect: {
        delay: 1000,
      },

      onConnected: () => {
        if (wasConnected) {
          livestreamAlert.value = {
            title: "Livestream Reconnected",
            message: "Livestream connection restored",
            type: "success",
          };

          setTimeout(() => {
            livestreamAlert.value = null;
          }, 3000);
        }

        wasConnected = true;
        isDisconnectedAlertVisible = false;
      },

      onDisconnected: (ws, event) => {
        if (!wasConnected) return;

        if (isDisconnectedAlertVisible) return;

        isDisconnectedAlertVisible = true;

        const closeEvent = event as CloseEvent;

        console.log("livestream disconnected, code:", closeEvent);

        if (closeEvent?.code === 1011) {
          livestreamAlert.value = {
            title: "Livestream Error",
            message: closeEvent.reason || "Livestream subscription failed",
            type: "error",
          };
        } else {
          livestreamAlert.value = {
            title: "Livestream Disconnected",
            message: "Livestream connection lost",
            type: "warning",
          };
        }
      },
    });
  }

  return {
    livestreamWs,
    livestreamAlert,
  };
}
