import type { PerspectiveCamera } from "three";
import { TresCanvas } from "@tresjs/core";
import type { Reactive } from "vue";
import * as THREE from "three";

export const camera: Ref<PerspectiveCamera | null> = ref(null);

export const cameraPosition: Reactive<THREE.Vector3> = reactive(
  new THREE.Vector3(),
);

export const tresCanvasParent: Ref<HTMLDivElement | null> = ref(null);
