import { WorkspaceEventRequest } from "~/messages/protobufs/workspace_event";
import type { SceneStates } from "~/types/scene-states";
import { transformProtoEventToCamera } from "../scene-states-provider/create-scene-states";
import { transformProtoToFace } from "./use-autosave";

export function useLivestream(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  if (
    workspace == undefined ||
    workspace == "me" ||
    sceneStates.livestreamWebsocket == undefined
  ) {
    return;
  }

  function handleWorkspaceEvent(resp: WorkspaceEventRequest) {
    for (const event of resp.autosave?.events ?? []) {
      if (event.calibrate != undefined) {
        sceneStates.calibration.heightOffset = event.calibrate.modelHeight;
        sceneStates.calibration.scale = event.calibrate.scaleFactor;
      } else if (event.delete != undefined) {
        delete sceneStates.cameras[event.delete.id];
      } else if (event.faceDelete != undefined) {
        delete sceneStates.facesManagement.faces[event.faceDelete.id];
      } else if (
        event.faceUpsert != undefined &&
        event.faceUpsert.coverageFace != undefined
      ) {
        const face = event.faceUpsert.coverageFace;
        sceneStates.facesManagement.faces[face.id] = transformProtoToFace(face);
      } else if (
        event.upsert != undefined &&
        event.upsert.camera != undefined
      ) {
        sceneStates.cameras[event.upsert.camera.id] =
          transformProtoEventToCamera(event.upsert.camera);
      }
    }
  }

  onMounted(() => {
    watch(
      () => sceneStates.livestreamWebsocket?.data.value,
      async (messageBlob) => {
        if (!messageBlob) return;
        const buf = await (messageBlob as Blob).arrayBuffer();
        const resp = WorkspaceEventRequest.decode(new Uint8Array(buf));

        handleWorkspaceEvent(resp);
      },
    );
  });
}
