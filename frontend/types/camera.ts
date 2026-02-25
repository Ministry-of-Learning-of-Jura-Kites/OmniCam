import { Vector3, Euler } from "three";
import {
  DistortionMode,
  type ColorRGBA,
} from "~/messages/protobufs/autosave_event";

export interface ICamera {
  name: string;
  position: Vector3;
  rotation: Euler;
  fov: number;
  aspectWidth: number;
  aspectHeight: number;
  frustumColor: ColorRGBA;
  frustumLength: number;
  distortion: {
    intensity: number;
    mode: DistortionMode;
  };
  isHidingArrows: boolean;
  isHidingWheels: boolean;
  isHidingFrustum: boolean;
  isLockingPosition: boolean;
  isLockingRotation: boolean;
  controlling?: {
    type: "rotation" | "moving";
    direction: "x" | "y" | "z";
  };
}

export const cameraDefault: ICamera = {
  name: "",
  rotation: new Euler(),
  position: new Vector3(),
  fov: 60,
  aspectWidth: 4,
  aspectHeight: 3,
  isHidingArrows: false,
  isHidingWheels: false,
  isLockingPosition: false,
  isLockingRotation: false,
  distortion: {
    intensity: 1,
    mode: DistortionMode.NONE,
  },
  isHidingFrustum: true,
  controlling: undefined,
  frustumColor: { r: 1, g: 0.8, b: 0.2, a: 0.5 },
  frustumLength: 10,
};
