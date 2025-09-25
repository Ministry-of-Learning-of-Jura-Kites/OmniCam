import type { ICamera } from "~/types/camera";

export function exportCamerasToJson(cameras: Record<string, ICamera>) {
  const data = Object.entries(cameras).map(([id, camera]) => ({
    id,
    name: camera.name,
    fov: camera.fov,
    isHidingArrows: camera.isHidingArrows,
    isHidingWheels: camera.isHidingWheels,
    controlling: camera.controlling,
    position: camera.position.toArray(),
    rotation: {
      x: camera.rotation.x,
      y: camera.rotation.y,
      z: camera.rotation.z,
      order: camera.rotation.order,
    },
  }));

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
