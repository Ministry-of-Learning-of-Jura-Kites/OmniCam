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
      fov: 60,
    };
    return camId;
  }

  function getCams() {
    return sceneStates.cameras;
  }

  function switchToCam(camId: string) {
    const cam = sceneStates.cameras[camId];
    sceneStates.currentCam.value = camId;
    gsap.to(sceneStates.tresContext.value!.camera!.position!, {
      x: cam?.position.x,
      y: cam!.position.y,
      z: cam!.position.z,
      onComplete: () => {
        sceneStates.currentCameraPosition.value = cam!.position;
      },
    });
    gsap.to(sceneStates.tresContext.value!.camera!.rotation!, {
      x: cam?.rotation.x,
      y: cam!.rotation.y,
      z: cam!.rotation.z,
      onComplete: () => {
        sceneStates.currentCameraRotation.value = cam!.rotation;
      },
    });
  }

  function switchToSpectator() {
    const camId = sceneStates.currentCam.value;
    const cam = sceneStates.cameras[camId!];
    const threeCam = sceneStates.tresContext.value!.camera!;
    cam?.position.copy(threeCam.position);
    cam?.rotation.copy(threeCam.rotation);
    gsap.to(threeCam.position!, {
      x: sceneStates.spectatorCameraPosition.x,
      y: sceneStates.spectatorCameraPosition.y,
      z: sceneStates.spectatorCameraPosition.z,
      onComplete: () => {
        sceneStates.currentCameraPosition.value =
          sceneStates.spectatorCameraPosition;
      },
    });
    gsap.to(threeCam!.rotation!, {
      x: sceneStates.spectatorCameraPosition.x,
      y: sceneStates.spectatorCameraPosition.y,
      z: sceneStates.spectatorCameraPosition.z,
      onComplete: () => {
        sceneStates.currentCameraRotation.value =
          sceneStates.spectatorCameraRotation;
      },
    });
    sceneStates.currentCam.value = null;
  }
  return {
    spawnCameraHere,
    switchToCam,
    getCams,
    switchToSpectator,
  };
}

export type CameraManament = ReturnType<typeof useCameraManagement>;
