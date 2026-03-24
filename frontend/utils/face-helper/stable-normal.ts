import { WORLD_UP } from "~/constants/three";
import { averageVector } from "./avg-vec";
import { Vector3 } from "three";

export function computeStableNormal(points: Vector3[], normals: Vector3[]) {
  const avgNormal = averageVector(normals);
  if (avgNormal.lengthSq() > 1e-8) return avgNormal.normalize();

  if (points.length >= 3) {
    const a = points[0]!;
    const b = points[1]!;
    const c = points[2]!;
    const n = new Vector3()
      .subVectors(b, a)
      .cross(new Vector3().subVectors(c, a));

    if (n.lengthSq() > 1e-8) return n.normalize();
  }

  return WORLD_UP.clone();
}
