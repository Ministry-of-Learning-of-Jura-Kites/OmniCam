import type { ICamera } from "~/types/camera";
import type { SceneStates } from "~/types/scene-states";

function isEqual(a: ICamera, b: ICamera): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export type AutosaveEvent =
  | { type: "delete"; data: string }
  | { type: "upsert"; data: ICamera };

export function useAutosave(sceneStates: SceneStates) {
  const lastSynced: Map<string, ICamera> = new Map(
    Object.entries(sceneStates.cameras!),
  );

  const markedForCheck = new Set<string>();

  // detect changes batchly
  watch(
    sceneStates.cameras,
    (newVal, oldVal) => {
      for (const [camId, cam] of Object.entries(newVal)) {
        const prevCam = oldVal[camId];
        if (prevCam == undefined || !isEqual(prevCam, cam)) {
          markedForCheck.add(camId);
        }
      }
    },
    { deep: true },
  );

  setInterval(() => {
    const newVal = sceneStates.cameras;
    const changed: AutosaveEvent[] = [];

    // check new or updated
    for (const camId of markedForCheck) {
      const prev = lastSynced.get(camId);
      const cam = newVal[camId];
      if (cam == undefined && prev == undefined) {
        continue;
      }
      // If is deleted
      if (cam == undefined) {
        lastSynced.delete(camId);
        changed.push({ type: "delete", data: camId });
        continue;
      }
      // If is newly added, or changed
      if (prev == undefined || !isEqual(prev, cam)) {
        changed.push({ type: "upsert", data: cam });
        lastSynced.set(camId, structuredClone(cam));
      }
    }

    markedForCheck.clear();
  }, 1000);
}
