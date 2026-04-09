import {
  WorkspaceEventRequest,
  WorkspaceEventResponse,
} from "~/messages/protobufs/workspace_event";
import type { SceneStates } from "~/types/scene-states";

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
      } else if (event.delete != undefined) {
      } else if (event.faceDelete != undefined) {
      } else if (event.faceUpsert != undefined) {
      } else if (event.upsert != undefined) {
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
