import { WorkspaceEventResponse } from "~/messages/protobufs/workspace_event";
import type { SceneStates } from "~/types/scene-states";

export function useLivestream(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  if (
    workspace == undefined ||
    workspace == "me" ||
    sceneStates.livestreamSse == undefined
  ) {
    return;
  }

  function handleWorkspaceEvent(resp: WorkspaceEventResponse) {}

  onMounted(() => {
    watch(
      () => sceneStates.websocket?.data.value,
      async (messageBlob) => {
        if (!messageBlob) return;
        const buf = await (messageBlob as Blob).arrayBuffer();
        const resp = WorkspaceEventResponse.decode(new Uint8Array(buf));

        handleWorkspaceEvent(resp);
      },
    );
  });
}
