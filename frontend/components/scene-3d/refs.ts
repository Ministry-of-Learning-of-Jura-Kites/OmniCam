import type { PerspectiveCamera } from "three";
import type { Reactive } from "vue";
import * as THREE from "three";

export const camera: Ref<PerspectiveCamera | null> = ref(null);

export const cameraPosition: Reactive<THREE.Vector3> = reactive(
  new THREE.Vector3(),
);

export const cameraRotation: Reactive<THREE.Euler> = reactive(
  new THREE.Euler(0, 0, 0, "YXZ"),
);

export const tresCanvasParent: Ref<HTMLDivElement | null> = ref(null);
