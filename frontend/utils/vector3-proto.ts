import type { Vector3 } from "~/messages/protobufs/autosave_event";

export function vector3ToNumbers(vec: Vector3) {
  return [vec.x, vec.y, vec.z];
}

export function numbersToVector3(vec: [number, number, number]) {
  return {
    x: vec[0],
    y: vec[1],
    z: vec[2],
  };
}
