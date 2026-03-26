import type { SceneStates } from "~/types/scene-states";
import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";
import { transformFaceToProto } from "./use-autosave";
import type { CameraConfig } from "~/messages/protobufs/optimization";
import { WorkspaceEventRequest } from "~/messages/protobufs/workspace_event";
import type { ICamera } from "~/types/camera";

// type OptimizationCallback = (opt: OptimizationEventResp) => void;

export function useOptimize(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  if (workspace != "me") {
    return null;
  }

  const candidateCameras = reactive<Record<string, ICamera>>({});
  const submitStatus = ref<"idle" | "sending" | "optimizing">("idle");

  function requestOptimize(
    targetAreaEntries: [string, ProcessedCoverageFace][],
    scale: number,
    cameraConfigs: CameraConfig[],
  ): boolean {
    if (!sceneStates.websocket) return false;

    const encoded = WorkspaceEventRequest.encode({
      optimize: {
        coverageFace: targetAreaEntries.map(([id, face]) =>
          transformFaceToProto(id, face),
        ),
        cameraConfig: cameraConfigs,
        scale: scale,
      },
    }).finish();
    sceneStates.websocket.send(encoded.buffer);
    submitStatus.value = "sending";

    return true;
  }

  return {
    requestOptimize,
    candidateCameras,
    submitStatus,
  };
}
