import type { TresContext } from "@tresjs/core";
import type { Reactive } from "vue";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import type {
  SceneStates as BaseSceneStates,
  SceneStatesWithHelper,
} from "~/types/scene-states";
import * as THREE from "three";
import type { ICamera } from "~/types/camera";
import { useCameraManagement } from "../scene-3d/use-camera-management";
import { useSpectatorRotation } from "../scene-3d/use-spectator-rotation";
import { useSpectatorPosition } from "../scene-3d/use-spectator-position";

export const SCENE_STATES_KEY: InjectionKey<SceneStatesWithHelper> =
  Symbol("3d-scene-states");

export function createBaseSceneStates() {
  const tresContext = ref<TresContext | null>(null);

  const draggableObjects: Set<Obj3DWithUserData> = new Set();

  const isDraggingObject: Ref<boolean> = ref(false);

  const currentCamId: Ref<string | null> = ref(null);

  const spectatorCameraPosition: Reactive<THREE.Vector3> = reactive(
    new THREE.Vector3(),
  );

  const spectatorCameraRotation: Reactive<THREE.Euler> = reactive(
    new THREE.Euler(0, 0, 0, "YXZ"),
  );

  const spectatorCameraFov: Ref<number> = ref(75);

  const currentCameraPosition: Ref<THREE.Vector3> = ref(
    spectatorCameraPosition,
  );

  const currentCameraRotation: Ref<THREE.Euler> = ref(spectatorCameraRotation);

  const tresCanvasParent: Ref<HTMLDivElement | null> = ref(null);

  const cameras = reactive<Record<string, ICamera>>({});

  const currentCam = computed(() => {
    return currentCamId.value == null ? null : cameras![currentCamId.value];
  });

  const currentCameraFov = computed({
    get() {
      if (currentCamId.value == null) {
        return spectatorCameraFov.value;
      }
      return currentCam.value?.fov;
    },
    set(newFOV) {
      if (currentCam.value) {
        currentCam.value.fov = newFOV!;
      }
    },
  });

  const sceneStates = {
    tresContext,
    draggableObjects,
    isDraggingObject,
    currentCamId,
    currentCam,
    currentCameraPosition,
    currentCameraRotation,
    currentCameraFov,
    spectatorCameraPosition,
    spectatorCameraRotation,
    spectatorCameraFov,
    tresCanvasParent,
    cameras,
  };

  return sceneStates;
}

export function createSceneStatesWithHelper(sceneStates: BaseSceneStates) {
  const sceneStatesWithCam = {
    ...sceneStates,
    cameraManagement: useCameraManagement(sceneStates),
    spectatorPosition: useSpectatorPosition(sceneStates),
    spectatorRotation: useSpectatorRotation(sceneStates),
  };
  return sceneStatesWithCam;
}
