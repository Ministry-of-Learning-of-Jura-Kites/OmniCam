import * as THREE from "three";
import type { ColorRGBA } from "~/messages/protobufs/autosave_event";

export interface ICamera {
  name: string;
  position: THREE.Vector3;
  rotation: THREE.Euler;
  fov: number;
  aspectWidth: number;
  aspectHeight: number;
  frustumColor: ColorRGBA;
  frustumLength: number;
  isHidingArrows: boolean;
  isHidingWheels: boolean;
  isLockingPosition: boolean;
  isLockingRotation: boolean;
  isHidingFrustum: boolean;
  controlling?: {
    type: "rotation" | "moving";
    direction: "x" | "y" | "z";
  };
}

export const cameraDefault: ICamera = {
  name: "",
  rotation: new THREE.Euler(),
  position: new THREE.Vector3(),
  fov: 60,
  aspectWidth: 4,
  aspectHeight: 3,
  isHidingArrows: false,
  isHidingWheels: false,
  isLockingPosition: false,
  isLockingRotation: false,
  isHidingFrustum: true,
  controlling: undefined,
  frustumColor: { r: 1, g: 0.8, b: 0.2, a: 0.5 },
  frustumLength: 0,
};
