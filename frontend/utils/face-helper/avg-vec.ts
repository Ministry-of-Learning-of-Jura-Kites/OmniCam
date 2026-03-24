import { Vector3 } from "three";

export function averageVector(vs: Vector3[]) {
  const out = new Vector3();
  if (vs.length === 0) return out;
  vs.forEach((v) => out.add(v));
  out.multiplyScalar(1 / vs.length);
  return out;
}
