import { v4 as uuidv4 } from "uuid";
import { gsap } from "gsap";
import type { SceneStates } from "~/types/scene-states";
import { Vector3, Euler } from "three";
import { cameraDefault, type ICamera } from "~/types/camera";
import { randomVividColor } from "~/utils/randomVividColor";
export function useCameraManagement(sceneStates: SceneStates) {
  function spawnCameraHere() {
    const camId = uuidv4();
    sceneStates.cameras[camId] = {
      ...cameraDefault,
      name: "Untitled " + Object.keys(sceneStates.cameras).length.toString(),
      position: new Vector3().copy(sceneStates.spectatorCameraPosition),
      rotation: new Euler().copy(sceneStates.spectatorCameraRotation),
      fov: 60,
      frustumColor: randomVividColor(),
    };
    return camId;
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
    await interpolateTransform(cam);
    sceneStates.currentCamId.value = camId;
    sceneStates.transformingInfo.value = undefined;
  }

  async function interpolateTransform(cam: ICamera) {
    if (sceneStates.transformingInfo.value == undefined) {
      return;
    }
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
  }

  async function switchToSpectator() {
    // Use currentCam to support teleporting from spectator to sp
    const cam = sceneStates.currentCam.value!;
    sceneStates.transformingInfo.value = {
      position: cam.position.clone(),
      rotation: cam.rotation.clone(),
      fov: cam?.fov,
    };

    await interpolateTransform(sceneStates.spectatorCam);

    sceneStates.currentCamId.value = null;
    sceneStates.transformingInfo.value = undefined;
  }
  return {
    spawnCameraHere,
    switchToCam,
    getCams,
    interpolateTransform,
    switchToSpectator,
  };
}

export type CameraManament = ReturnType<typeof useCameraManagement>;
