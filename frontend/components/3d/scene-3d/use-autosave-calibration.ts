import { useWebSocket } from "@vueuse/core";
import {
  CalibrationSaveEvent,
  CalibrationAutosaveResponse,
} from "~/messages/protobufs/autosave_event";
import type { SceneStates } from "~/types/scene-states";

export function useAutosaveCalibration(
  sceneStates: SceneStates,
  workspace: string | null,
  projectId: string,
  modelId: string,
) {
  if (workspace == null) return;

  const runtimeConfig = useRuntimeConfig();

  const wsUrl = `ws://${runtimeConfig.public.externalBackendHost}/api/v1/projects/${projectId}/models/${modelId}/autosave/calibration`;

  const ws = useWebSocket(wsUrl, {
    onConnected(socket) {
      socket.binaryType = "blob";
    },
    autoReconnect: {
      delay: 1000,
      onFailed: () => {
        alert(
          "Failed to connect calibration websocket after multiple retries.",
        );
      },
    },
  });

  watch(
    () => [
      sceneStates.calibrationScale.value,
      sceneStates.calibrationHeight.value,
    ],
    ([newScale, newHeight], [oldScale, oldHeight]) => {
      // Ignore changes coming FROM the server ACK (would cause infinite loop)
      if (newScale !== oldScale || newHeight !== oldHeight) {
        sceneStates.calibrationDirty.value = true;
      }
    },
  );

  // Receive: initial values on connect + ACKs after each save
  watch(
    () => ws.data.value,
    async (blob) => {
      if (!blob) return;
      const buf = await (blob as Blob).arrayBuffer();
      const resp = CalibrationAutosaveResponse.decode(new Uint8Array(buf));
      if (resp.ack) {
        sceneStates.calibrationScale.value = resp.ack.scaleFactor;
        sceneStates.calibrationHeight.value = resp.ack.modelHeight;
        sceneStates.calibrationVersion.value = resp.ack.lastUpdatedVersion;
      }
    },
  );

  // Send: every 2s if dirty, same pattern as use-autosave.ts
  setInterval(() => {
    if (!sceneStates.calibrationDirty.value) return;

    sceneStates.calibrationVersion.value += 1;

    const encoded = CalibrationSaveEvent.encode({
      version: sceneStates.calibrationVersion.value,
      calibration: {
        scaleFactor: sceneStates.calibrationScale.value,
        modelHeight: sceneStates.calibrationHeight.value,
      },
    }).finish();
    ws.send(encoded.buffer);
    sceneStates.calibrationDirty.value = false;
  }, 2000);
}
