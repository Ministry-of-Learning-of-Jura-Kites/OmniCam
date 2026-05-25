export const BASE64_UUID_SIZE = 22;

export function uuidToBase64Url(uuidString: string) {
  // Remove hyphens and convert to a continuous hex string
  const hex = uuidString.replaceAll(/=/g, "");

  // Convert hex string to a Uint8Array
  const bytes = new Uint8Array(16);
  for (let i = 0; i < 16; i++) {
    bytes[i] = Number.parseInt(hex.substring(i * 2, i * 2 + 2), 16);
  }

  // Convert Uint8Array to a binary string (each byte as a character)
  const base64Standard = btoa(String.fromCodePoint(...bytes));

  // Base64 encode the binary string
  return base64Standard
    .replaceAll(/\+/g, "-")
    .replaceAll(/\//g, "_")
    .replaceAll(/=/g, "");
}
