import type * as THREE from "three";

export interface ICamera {
  // id: string;
  name: string;
  // color: ;
  position: THREE.Vector3;
  rotation: THREE.Euler;
  fov: number;
}
