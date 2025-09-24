import type { ClassValue } from "clsx";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import * as THREE from "three";

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
