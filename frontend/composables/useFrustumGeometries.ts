import { ref, onUnmounted } from "vue";
import * as THREE from "three";
import { createFrustumGeometry } from "~/components/3d/camera-frustum/create-frustum";

const frustumRegistry = ref<
  Record<string, { mesh: THREE.BufferGeometry; lines: THREE.BufferGeometry }>
>({});

export function useFrustumGeometries() {
  function setFrustumGeometry(
    id: string,
    fov: number,
    aspect: number,
    length: number,
  ) {
    const old = frustumRegistry.value[id];
    if (old) {
      old.mesh.dispose();
      old.lines.dispose();
    }

    const pair = createFrustumGeometry(fov, aspect, length);
    frustumRegistry.value[id] = pair;
    return pair;
  }

  function removeFrustumGeometry(id: string) {
    const geom = frustumRegistry.value[id];
    if (geom) {
      geom.mesh.dispose();
      geom.lines.dispose();
      delete frustumRegistry.value[id];
    }
  }

  function getFrustumGeometry(id: string) {
    return frustumRegistry.value[id] ?? null;
  }

  function getAllFrustums() {
    return frustumRegistry.value;
  }

  return {
    setFrustumGeometry,
    removeFrustumGeometry,
    getFrustumGeometry,
    getAllFrustums,
  };
}
