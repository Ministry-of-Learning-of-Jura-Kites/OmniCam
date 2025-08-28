import type { TresContext } from "@tresjs/core";
import type { Reactive } from "vue";
import type { ObjWithUserData } from "./obj-event-handler";
import * as THREE from "three";

interface Camera {
  // id: string;
  name: string;
  // color: ;
  position: THREE.Vector3;
  rotation: THREE.Euler;
}

export const SCENE_STATES_KEY: InjectionKey<SceneStates> =
  Symbol("3d-scene-states");

export function createSceneStates() {
  const tresContext = ref<TresContext | null>(null);

  const draggableObjects: Set<ObjWithUserData> = new Set();

  const isDraggingObject: Ref<boolean> = ref(false);

  const currentCam: Ref<string | null> = ref(null);

  const spectatorCameraPosition: Reactive<THREE.Vector3> = reactive(
    new THREE.Vector3(),
  );

  const spectatorCameraRotation: Reactive<THREE.Euler> = reactive(
    new THREE.Euler(0, 0, 0, "YXZ"),
  );
  const currentCameraPosition: Ref<THREE.Vector3> = ref(
    spectatorCameraPosition,
  );

  const currentCameraRotation: Ref<THREE.Euler> = ref(spectatorCameraRotation);

  const tresCanvasParent: Ref<HTMLDivElement | null> = ref(null);

  const cameras = reactive<Record<string, Camera>>({});

  return {
    tresContext,
    draggableObjects,
    isDraggingObject,
    currentCam,
    currentCameraPosition,
    currentCameraRotation,
    spectatorCameraPosition,
    spectatorCameraRotation,
    tresCanvasParent,
    cameras,
  };
}

export type SceneStates = ReturnType<typeof createSceneStates>;
