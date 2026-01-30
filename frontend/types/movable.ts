import type { Vector3 } from "three";

export interface MovableObject {
  position: Vector3;
  controlling?: {
    type: "rotation" | "moving";
    direction: "x" | "y" | "z";
  };
  isHidingArrows?: boolean;
}
