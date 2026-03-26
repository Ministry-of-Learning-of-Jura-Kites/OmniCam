export type FixedLengthArray<
  TItem,
  TLength extends number,
  TAcc extends TItem[] = [],
> = TAcc["length"] extends TLength
  ? TAcc
  : FixedLengthArray<TItem, TLength, [...TAcc, TItem]>;

export type Trapezoid = FixedLengthArray<[number, number, number], 4>;
