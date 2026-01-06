import { Vector3 } from "three";

export function getAxisVector(axis: "x" | "y" | "z") {
  return new Vector3(
    axis === "x" ? 1 : 0,
    axis === "y" ? 1 : 0,
    axis === "z" ? 1 : 0,
  );
}
