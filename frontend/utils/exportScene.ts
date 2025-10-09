import type { ICamera } from "~/types/camera";
import * as THREE from "three";

export function exportCamerasToJson(cameras: Record<string, ICamera>) {
  const data = Object.entries(cameras).map(([id, camera]) => {
    const quat = new THREE.Quaternion().setFromEuler(camera.rotation);

    return {
      id,
      Name: camera.name,
      AngleX: quat.x,
      AngleY: quat.y,
      AngleZ: quat.z,
      AngleW: quat.w,
      PosX: camera.position.x,
      PosY: camera.position.y,
      PosZ: camera.position.z,
      Fov: camera.fov,
      IsHidingArrows: camera.isHidingArrows,
      IsHidingWheels: camera.isHidingWheels,
      IsLockingPosition: camera.isLockingPosition,
      IsLockingRotation: camera.isLockingRotation,
    };
  });

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
