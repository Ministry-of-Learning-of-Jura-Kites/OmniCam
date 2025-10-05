import type { ClassValue } from "clsx";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import * as THREE from "three";
import "uuid";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function getAxisVector(axis: "x" | "y" | "z") {
  return new THREE.Vector3(
    axis === "x" ? 1 : 0,
    axis === "y" ? 1 : 0,
    axis === "z" ? 1 : 0,
  );
}

export function uuidToBase64Url(uuidString: string) {
  // Remove hyphens and convert to a continuous hex string
  const hex = uuidString.replace(/-/g, "");

  // Convert hex string to a Uint8Array
  const bytes = new Uint8Array(16);
  for (let i = 0; i < 16; i++) {
    bytes[i] = parseInt(hex.substring(i * 2, i * 2 + 2), 16);
  }

  // Convert Uint8Array to a binary string (each byte as a character)
  const base64Standard = btoa(String.fromCharCode(...bytes));

  // Base64 encode the binary string
  return base64Standard
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}
