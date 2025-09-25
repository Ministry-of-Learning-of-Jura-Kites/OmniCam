import type { ICamera } from "~/types/camera";
import * as THREE from "three";

export function importJsonToCameras(
  sceneCameras: Record<string, ICamera>,
  jsonData: string,
) {
  try {
    const data = JSON.parse(jsonData) as Array<{
      id: string;
      Name: string;
      AngleX: number;
      AngleY: number;
      AngleZ: number;
      AngleW: number;
      PosX: number;
      PosY: number;
      PosZ: number;
      Fov: number;
      IsHidingArrows: boolean;
      IsHidingWheels: boolean;
    }>;

    const newCameras: Record<string, ICamera> = {};

    for (const camera of data) {
      newCameras[camera.id] = {
        name: camera.Name,
        fov: camera.Fov,
        isHidingArrows: camera.IsHidingArrows,
        isHidingWheels: camera.IsHidingWheels,
        controlling: null,
        position: new THREE.Vector3(camera.PosX, camera.PosY, camera.PosZ),
        rotation: new THREE.Euler().setFromQuaternion(
          new THREE.Quaternion(
            camera.AngleX,
            camera.AngleY,
            camera.AngleZ,
            camera.AngleW,
          ),
        ),
      };
    }

    Object.keys(sceneCameras).forEach((key) => {
      Reflect.deleteProperty(sceneCameras, key);
    });

    Object.assign(sceneCameras, newCameras);
  } catch (err) {
    console.error("Failed to import cameras JSON:", err);
  }
}
