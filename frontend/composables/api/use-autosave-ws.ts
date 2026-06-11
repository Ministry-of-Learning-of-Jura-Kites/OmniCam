import { useWebSocket, type UseWebSocketReturn } from "@vueuse/core";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export function useAutosaveWs(
  projectId: string,
  modelId: string,
  workspace: string,
) {
  const { $otel } = useNuxtApp();
  const runtimeConfig = useRuntimeConfig();
  let autosaveWs: UseWebSocketReturn<unknown> | undefined = undefined;
  const autosaveAlert = ref<{
    title: string;
    message: string;
    type: "success" | "warning" | "error";
  } | null>(null);
  if (workspace == "me" && import.meta.client) {
    const base = getApiBaseUrlWithProtocol("websocket", runtimeConfig);
    const websocketUrl = concatUrl(
      `projects/${projectId}/models/${modelId}/autosave`,
      base,
    );

    let wasConnected = false;
    let isDisconnectedAlertVisible = false;

    autosaveWs = useWebSocket(websocketUrl, {
      autoReconnect: {
        delay: 1000,
      },

      onConnected: () => {
        $otel.traceWsSend(autosaveWs?.ws.value, "connect");
        if (wasConnected) {
          autosaveAlert.value = {
            title: "Autosave Reconnected",
            message: "Autosave connection restored",
            type: "success",
          };

          setTimeout(() => {
            autosaveAlert.value = null;
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

        if (closeEvent?.code === 1011) {
          autosaveAlert.value = {
            title: "Autosave Error",
            message: closeEvent.reason || "Autosave subscription failed",
            type: "error",
          };
        } else {
          autosaveAlert.value = {
            title: "Autosave Disconnected",
            message: "Changes may not be saved until reconnect",
            type: "warning",
          };
        }
      },
      onMessage: $otel.traceWsReceive(
        websocketUrl.toString(),
        "autosave",
        (data) => {
          console.log("[Autosave] message received", data);
        },
      ),
    });
  }

  return {
    autosaveWs,
    autosaveAlert,
  };
}
