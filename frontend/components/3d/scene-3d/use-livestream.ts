import { LivestreamBroadcast } from "~/messages/protobufs/workspace_event";
import type { SceneStates } from "~/types/scene-states";
import { transformProtoEventToCamera } from "../scene-states-provider/create-scene-states";
import { transformProtoToFace } from "./use-autosave";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

export function useLivestream(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  const sceneState = inject(SCENE_STATES_KEY);
  if (
    workspace == undefined ||
    workspace == "me" ||
    sceneStates.livestreamWebsocket == undefined
  ) {
    return;
  }

  function handleWorkspaceEvent(resp: LivestreamBroadcast) {
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
      } else {
        console.warn("unhandled livestream event:", event);
      }
    }
  }

  watch(
    () => sceneStates.livestreamWebsocket?.data.value,
    async (messageBlob) => {
      if (!messageBlob) return;
      const buf = await (messageBlob as Blob).arrayBuffer();
      const resp = LivestreamBroadcast.decode(new Uint8Array(buf));
      if (resp.error) {
        console.log("in 1");
        console.log(sceneState?.value);
        if (sceneState?.value) {
          console.log("in 2");
          sceneState.value.errorLivestreamMessage.value =
            "A livestream connection error occurred. Please try again.";
        }
        return;
      }

      handleWorkspaceEvent(resp);
    },
  );
}
