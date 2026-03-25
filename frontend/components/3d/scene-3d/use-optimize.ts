import type { SceneStates } from "~/types/scene-states";
import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";
import { transformFaceToProto } from "./use-autosave";
import type { CameraConfig } from "~/messages/protobufs/optimization";
import {
  WorkspaceEventResponse,
  WorkspaceEventRequest,
} from "~/messages/protobufs/workspace_event";

export function useOptimize(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  if (workspace != "me") {
    return null;
  }

  onMounted(() => {
    watch(
      () => sceneStates.websocket?.data.value,
      async (messageBlob) => {
        if (!messageBlob) return;
        const buf = await (messageBlob as Blob).arrayBuffer();
        const resp = WorkspaceEventResponse.decode(new Uint8Array(buf));

        if (resp.optimize) {
          console.log(resp.optimize);
        }
      },
    );
  });

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

    return true;
  }

  return {
    requestOptimize,
  };
}
