import { Vector3 } from "three";

export function threeVector3ToNumbers(vec: Vector3) {
  return [vec.x, vec.y, vec.z] as [number, number, number];
}

export function numbersToThreeVector3(vec: [number, number, number]) {
  return new Vector3(vec[0], vec[1], vec[2]);
}
