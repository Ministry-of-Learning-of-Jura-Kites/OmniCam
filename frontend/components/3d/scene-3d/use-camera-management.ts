import { v4 as uuidv4 } from "uuid";
import { gsap } from "gsap";
import type { SceneStates } from "~/types/scene-states";
import * as THREE from "three";
import { cameraDefault } from "~/types/camera";

export function useCameraManagement(sceneStates: SceneStates) {
  function spawnCameraHere() {
    const camId = uuidv4();
    sceneStates.cameras[camId] = {
      ...cameraDefault,
      name: "Untitled " + Object.keys(sceneStates.cameras).length.toString(),
      position: new THREE.Vector3().copy(sceneStates.spectatorCameraPosition),
      rotation: new THREE.Euler().copy(sceneStates.spectatorCameraRotation),
      fov: 60,
      frustumColor: randomVividColor(),
      frustumLength: 10,
    };
    sceneStates.markedForCheck.add(camId);
    console.log(sceneStates.cameras[camId].frustumColor);
    return camId;
  }

  function randomVividColor() {
    //preset
    const high = 0.85 + Math.random() * 0.15; // ~0.85–1.00
    const mid = 0.5 + Math.random() * 0.3; // ~0.5–0.8
    const low = 0.1 + Math.random() * 0.2; // ~0.1–0.3

    const patterns: [number, number, number][] = [
      [high, high, low],
      [high, mid, mid],
      [high, mid, low],
      [high, low, low],
    ];

    const chosen = patterns[Math.floor(Math.random() * patterns.length)];

    const shuffled = chosen!.sort(() => Math.random() - 0.5);

    return {
      r: shuffled[0],
      g: shuffled[1],
      b: shuffled[2],
      a: 0.5,
    };
  }

  function getCams() {
    return sceneStates.cameras;
  }

  async function switchToCam(camId: string) {
    if (sceneStates.currentCamId.value == camId) {
      return;
    }
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
