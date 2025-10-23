import type { ICamera } from "~/types/camera";
import { transformCameraToProtoEvent } from "~/components/3d/scene-3d/use-autosave";

export function exportCamerasToJson(cameras: Record<string, ICamera>) {
  const data = Object.fromEntries(
    Object.entries(cameras).map(([id, camera]) => {
      return [id, transformCameraToProtoEvent(camera)];
    }),
  );

  const blob = new Blob([JSON.stringify(data, null, 2)], {
    type: "application/json",
  });

  const url = URL.createObjectURL(blob);
  const downloadElement = document.createElement("a");
  downloadElement.href = url;
  downloadElement.download = "cameras.json";
  downloadElement.click();
  URL.revokeObjectURL(url);
}
