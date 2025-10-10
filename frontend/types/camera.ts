import type * as THREE from "three";

// export interface VectorObj {
//   x: number;
//   y: number;
//   z: number;
// }

// export function vectorToVectorObj(vec: THREE.Vector3) {
//   return {
//     x: vec.x,
//     y: vec.y,
//     z: vec.z,
//   };
// }
// export function eulerToVectorObj(vec: THREE.Euler) {
//   return {
//     x: vec.x,
//     y: vec.y,
//     z: vec.z,
//   };
// }

export interface ICamera {
  // id: string;
  name: string;
  // color: ;
  position: THREE.Vector3;
  rotation: THREE.Euler;
  fov: number;
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
