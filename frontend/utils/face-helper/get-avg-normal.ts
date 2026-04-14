import type {
  QuadrilateralPoints,
  QuadrilateralVectors,
} from "~/types/trapezoid";
import { Vector3 } from "three";

export function arrayPointsToNormal(points: QuadrilateralPoints) {
  const transPoints = points.map(numbersToThreeVector3) as QuadrilateralVectors;
  return pointsToNormal(transPoints);
}

export function pointsToNormal(points: QuadrilateralVectors) {
  const edge1 = new Vector3().subVectors(points[0], points[1]).normalize();
  const edge2 = new Vector3().subVectors(points[2], points[3]).normalize();
  const normal = new Vector3().crossVectors(edge1, edge2).normalize();

  return normal;
}
