export type FixedLengthArray<
  TItem,
  TLength extends number,
  TAcc extends TItem[] = [],
> = TAcc["length"] extends TLength
  ? TAcc
  : FixedLengthArray<TItem, TLength, [...TAcc, TItem]>;

export type Trapezoid = FixedLengthArray<[number, number, number], 4>;

export function getTrapezoidNormal(
  trapezoid: Trapezoid,
): [number, number, number] {
  const [p0, p1, , p3] = trapezoid;

  // Calculate two vectors sharing the same origin (p0)
  const v1 = [p1[0] - p0[0], p1[1] - p0[1], p1[2] - p0[2]] as const;

  const v2 = [p3[0] - p0[0], p3[1] - p0[1], p3[2] - p0[2]] as const;

  // Cross Product calculation
  // nx = (ay * bz) - (az * by)
  // ny = (az * bx) - (ax * bz)
  // nz = (ax * by) - (ay * bx)
  const nx = v1[1] * v2[2] - v1[2] * v2[1];
  const ny = v1[2] * v2[0] - v1[0] * v2[2];
  const nz = v1[0] * v2[1] - v1[1] * v2[0];

  // Normalize
  const length = Math.sqrt(nx * nx + ny * ny + nz * nz);

  if (length === 0) {
    throw new Error(
      "Degenerate trapezoid: points are collinear or overlapping.",
    );
  }

  return [nx / length, ny / length, nz / length];
}
