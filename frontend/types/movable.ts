import type { Vector3, Euler } from "three";

export interface MovableObject {
  position: Vector3;
  rotation: Euler;
  controlling?: {
    type: "rotation" | "moving";
    direction: "x" | "y" | "z";
  };
  isHidingArrows?: boolean;
}
