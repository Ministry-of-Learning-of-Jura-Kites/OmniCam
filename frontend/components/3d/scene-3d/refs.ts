import type { Reactive } from "vue";
import * as THREE from "three";
import type { TresContext } from "@tresjs/core";
import type { IUserData } from "./obj-event-handler";

export const tresContext = ref<TresContext | null>(null);

export type ObjWithUserData = THREE.Object3D & { userData: IUserData };

export const draggableObjects: Set<ObjWithUserData> = new Set();

export const isDraggingObject: Ref<boolean> = ref(false);

export const currentCam: Ref<string | null> = ref(null);

export const spectatorCameraPosition: Reactive<THREE.Vector3> = reactive(
  new THREE.Vector3(),
);

export const spectatorCameraRotation: Reactive<THREE.Euler> = reactive(
  new THREE.Euler(0, 0, 0, "YXZ"),
);
export const currentCameraPosition: Ref<THREE.Vector3> = ref(
  spectatorCameraPosition,
);

export const currentCameraRotation: Ref<THREE.Euler> = ref(
  spectatorCameraRotation,
);

export const tresCanvasParent: Ref<HTMLDivElement | null> = ref(null);
