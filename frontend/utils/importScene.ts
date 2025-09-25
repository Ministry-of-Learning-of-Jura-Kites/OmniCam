import type { ICamera } from "~/types/camera";
import * as THREE from "three";

export function importJsonToCameras(
  sceneCameras: Record<string, ICamera>,
  jsonData: string,
) {
  try {
    const data = JSON.parse(jsonData) as Array<{
      id: string;
      name: string;
      fov: number;
      isHidingArrows: boolean;
      isHidingWheels: boolean;
      controlling: ICamera["controlling"];
      position: [number, number, number];
      rotation: { x: number; y: number; z: number; order: string };
    }>;

    const newCameras: Record<string, ICamera> = {};

    for (const camera of data) {
      newCameras[camera.id] = {
        name: camera.name,
        fov: camera.fov,
        isHidingArrows: camera.isHidingArrows,
        isHidingWheels: camera.isHidingWheels,
        controlling: camera.controlling,
        position: new THREE.Vector3(...camera.position),
        rotation: new THREE.Euler(
          camera.rotation.x,
          camera.rotation.y,
          camera.rotation.z,
          camera.rotation.order as THREE.EulerOrder,
        ),
      };
    }

    Object.assign(sceneCameras, {});
    Object.assign(sceneCameras, newCameras);
  } catch (err) {
    console.error("Failed to import cameras JSON:", err);
  }
}
