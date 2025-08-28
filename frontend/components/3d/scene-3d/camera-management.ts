import type * as THREE from "three";
import { v4 as uuidv4 } from "uuid";
import {
  currentCam,
  currentCameraPosition,
  currentCameraRotation,
  spectatorCameraPosition,
  spectatorCameraRotation,
  tresContext,
} from "./refs";
import { gsap } from "gsap";

interface Camera {
  // id: string;
  name: string;
  // color: ;
  position: THREE.Vector3;
  rotation: THREE.Euler;
}

export const cameras = reactive<Record<string, Camera>>({});

export function spawnCameraHere() {
  const camId = uuidv4();
  cameras[camId] = {
    name: "Untitled " + Object.keys(cameras).length.toString(),
    position: spectatorCameraPosition.clone(),
    rotation: spectatorCameraRotation.clone(),
  };
  return camId;
}

export function switchToCam(camId: string) {
  const cam = cameras[camId];
  currentCam.value = camId;
  gsap.to(tresContext.value!.camera!.position!, {
    x: cam?.position.x,
    y: cam!.position.y,
    z: cam!.position.z,
    onComplete: () => {
      currentCameraPosition.value = cam!.position;
    },
  });
  gsap.to(tresContext.value!.camera!.rotation!, {
    x: cam?.rotation.x,
    y: cam!.rotation.y,
    z: cam!.rotation.z,
    onComplete: () => {
      currentCameraRotation.value = cam!.rotation;
    },
  });
}

export function switchToSpectator() {
  const camId = currentCam.value;
  const cam = cameras[camId!];
  const threeCam = tresContext.value!.camera!;

  cam?.position.copy(threeCam.position);
  cam?.rotation.copy(threeCam.rotation);

  gsap.to(threeCam.position!, {
    x: spectatorCameraPosition.x,
    y: spectatorCameraPosition.y,
    z: spectatorCameraPosition.z,
    onComplete: () => {
      currentCameraPosition.value = spectatorCameraPosition;
    },
  });
  gsap.to(threeCam!.rotation!, {
    x: spectatorCameraRotation.x,
    y: spectatorCameraRotation.y,
    z: spectatorCameraRotation.z,
    onComplete: () => {
      currentCameraRotation.value = spectatorCameraRotation;
    },
  });
  currentCam.value = null;
}
