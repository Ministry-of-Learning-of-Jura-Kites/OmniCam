export function objectToQueryParams(
  obj: Record<string, number | string | string[] | number[] | Date>,
) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(obj)) {
    if (Array.isArray(value)) {
      for (const ele of value) {
        params.append(key, String(ele));
      }
    } else {
      params.append(key, String(value));
    }
  }
  return params;
}
