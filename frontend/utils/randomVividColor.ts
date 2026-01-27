export function randomVividColor() {
  //preset
  const high = 0.85 + Math.random() * 0.15; // ~0.85–1.00
  const mid = 0.5 + Math.random() * 0.3; // ~0.5–0.8
  const low = 0.1 + Math.random() * 0.2; // ~0.1–0.3

  const patterns: [number, number, number][] = [
    [high, high, low],
    [high, mid, mid],
    [high, mid, low],
    [high, low, low],
  ];

  const chosen = patterns[Math.floor(Math.random() * patterns.length)];

  const shuffled = chosen!.sort(() => Math.random() - 0.5);

  return {
    r: shuffled[0],
    g: shuffled[1],
    b: shuffled[2],
    a: 0.3,
  };
}
