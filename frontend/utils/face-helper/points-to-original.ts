import type { Vector3 } from "three";

export function matchPointsToOriginal(
  newPoints: Vector3[],
  originalPoints: Vector3[],
): Vector3[] {
  const result: Vector3[] = new Array(originalPoints.length);
  const remainingNew = [...newPoints];

  originalPoints.forEach((orig, i) => {
    // Find the point in the new set closest to where the original corner was
    let closestIdx = 0;
    let minDist = Infinity;

    remainingNew.forEach((p, j) => {
      const d = p.distanceTo(orig);
      if (d < minDist) {
        minDist = d;
        closestIdx = j;
      }
    });

    result[i] = remainingNew.splice(closestIdx, 1)[0]!;
  });

  return result;
}
