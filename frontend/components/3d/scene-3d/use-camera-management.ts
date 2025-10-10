import { v4 as uuidv4 } from "uuid";
import { gsap } from "gsap";
import type { SceneStates } from "~/types/scene-states";
import * as THREE from "three";

export function useCameraManagement(sceneStates: SceneStates) {
  function spawnCameraHere() {
    const camId = uuidv4();
    sceneStates.cameras[camId] = {
      name: "Untitled " + Object.keys(sceneStates.cameras).length.toString(),
      position: new THREE.Vector3().copy(sceneStates.spectatorCameraPosition),
      rotation: new THREE.Euler().copy(sceneStates.spectatorCameraRotation),
      isHidingArrows: false,
      isHidingWheels: false,
      isLockingPosition: false,
      isLockingRotation: false,
      controlling: undefined,
      fov: 60,
    };
    sceneStates.markedForCheck.add(camId);
    return camId;
  }

  function getCams() {
    return sceneStates.cameras;
  }

  async function switchToCam(camId: string) {
    const cam = sceneStates.cameras[camId]!;
    sceneStates.transformingInfo.value = {
      position: sceneStates.spectatorCameraPosition.clone(),
      rotation: sceneStates.spectatorCameraRotation.clone(),
      fov: sceneStates.spectatorCameraFov.value,
    };
    const tasks = [
      gsap.to(sceneStates.transformingInfo.value.position, {
        x: cam.position.x,
        y: cam.position.y,
        z: cam.position.z,
      }),
      gsap.to(sceneStates.transformingInfo.value!, {
        fov: cam?.fov,
      }),
      gsap.to(sceneStates.transformingInfo.value.rotation!, {
        x: cam.rotation.x,
        y: cam.rotation.y,
        z: cam.rotation.z,
      }),
    ];
    await Promise.all(tasks);
    sceneStates.currentCamId.value = camId;
    sceneStates.transformingInfo.value = undefined;
  }

  async function switchToSpectator() {
    const camId = sceneStates.currentCamId.value;
    const cam = sceneStates.cameras[camId!]!;
    sceneStates.transformingInfo.value = {
      position: cam.position.clone(),
      rotation: cam.rotation.clone(),
      fov: cam?.fov,
    };
    const tasks = [
      gsap.to(sceneStates.transformingInfo.value.position, {
        x: sceneStates.spectatorCameraPosition.x,
        y: sceneStates.spectatorCameraPosition.y,
        z: sceneStates.spectatorCameraPosition.z,
      }),
      gsap.to(sceneStates.transformingInfo.value!, {
        fov: sceneStates.spectatorCameraFov.value,
      }),
      gsap.to(sceneStates.transformingInfo.value.rotation!, {
        x: sceneStates.spectatorCameraRotation.x,
        y: sceneStates.spectatorCameraRotation.y,
        z: sceneStates.spectatorCameraRotation.z,
      }),
    ];
    await Promise.all(tasks);
    sceneStates.currentCamId.value = null;
    sceneStates.transformingInfo.value = undefined;
  }
  return {
    spawnCameraHere,
    switchToCam,
    getCams,
    switchToSpectator,
  };
}

export type CameraManament = ReturnType<typeof useCameraManagement>;
