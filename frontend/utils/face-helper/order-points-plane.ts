import { Vector3 } from "three";
import { WORLD_UP } from "~/constants/three";
import { averageVector } from "./avg-vec";

function makeBasisFromNormal(normal: Vector3) {
  const n = normal.clone().normalize();

  let u = new Vector3().crossVectors(WORLD_UP, n);
  if (u.lengthSq() < 1e-8) {
    u = new Vector3(1, 0, 0).cross(n);
  }
  if (u.lengthSq() < 1e-8) {
    u = new Vector3(0, 0, 1);
  }
  u.normalize();

  const v = new Vector3().crossVectors(n, u).normalize();
  return { u, v };
}

export function orderPointsOnPlane(points: Vector3[], normal: Vector3) {
  const center = averageVector(points);
  const { u, v } = makeBasisFromNormal(normal);

  const projected = points.map((p) => {
    const d = p.clone().sub(center);
    const du = d.dot(u);
    const dv = d.dot(v);
    return center.clone().addScaledVector(u, du).addScaledVector(v, dv);
  });

  return projected
    .map((p) => {
      const d = p.clone().sub(center);
      return {
        p,
        angle: Math.atan2(d.dot(v), d.dot(u)),
      };
    })
    .sort((a, b) => a.angle - b.angle)
    .map((x) => x.p);
}
