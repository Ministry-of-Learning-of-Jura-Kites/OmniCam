import type { SceneStates } from "~/types/scene-states";
import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";
import { transformFaceToProto } from "./use-autosave";
import type { CameraConfig } from "~/messages/protobufs/optimization";
import { WorkspaceEventRequest } from "~/messages/protobufs/workspace_event";
import type { ICamera } from "~/types/camera";
import { context, propagation, trace } from "@opentelemetry/api";

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
  const tracer = trace.getTracer("omnicam-frontend");

  function requestOptimize(
    targetAreaEntries: [string, ProcessedCoverageFace][],
    scale: number,
    cameraConfigs: CameraConfig[],
  ): boolean {
    if (!sceneStates.websocket) return false;

    const span = tracer.startSpan("websocket.send.optimize");

    try {
      const carrier: Record<string, string> = {};

      propagation.inject(trace.setSpan(context.active(), span), carrier);

      const encoded = WorkspaceEventRequest.encode({
        trace: {
          traceparent: carrier.traceparent ?? "",
          tracestate: carrier.tracestate ?? "",
        },
        optimize: {
          coverageFace: targetAreaEntries.map(([id, face]) =>
            transformFaceToProto(id, face),
          ),
          cameraConfig: cameraConfigs,
          scale,
        },
      }).finish();

      sceneStates.websocket.send(encoded.buffer);

      submitStatus.value = "sending";

      return true;
    } finally {
      span.end();
    }
  }

  return {
    requestOptimize,
    candidateCameras,
    submitStatus,
  };
}
