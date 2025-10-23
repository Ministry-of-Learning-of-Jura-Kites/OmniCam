import type { TresContext } from "@tresjs/core";
import type { Reactive } from "vue";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import type {
  SceneStates as BaseSceneStates,
  SceneStatesWithHelper,
} from "~/types/scene-states";
import * as THREE from "three";
import { cameraDefault, type ICamera } from "~/types/camera";
import { useCameraManagement } from "../scene-3d/use-camera-management";
import { useSpectatorRotation } from "../scene-3d/use-spectator-rotation";
import { useSpectatorPosition } from "../scene-3d/use-spectator-position";
import type { UseWebSocketReturn } from "@vueuse/core";
import type { Camera } from "~/messages/protobufs/autosave_event";
import { useAspectRatio as useAspectRatioManagement } from "../scene-3d/use-aspect-ratio";
import { useAutosave } from "../scene-3d/use-autosave";

export interface ModelWithCamsResp {
  data: {
    workspaceExists: boolean | null;
    modelId: string; // uuid.UUID -> string
    name: string;
    description: string;
    version: number; // int32 -> number
    createdAt: string;
    updatedAt: string;
    projectId: string; // uuid.UUID -> string
    filePath: string;
    imagePath: string;
    cameras: Record<string, Camera>;
  };
}

export const SCENE_STATES_KEY: InjectionKey<SceneStatesWithHelper> =
  Symbol("3d-scene-states");

export function transformProtoEventToCamera(rawCam: Camera): ICamera {
  return {
    name: rawCam.name,
    position: new THREE.Vector3(rawCam.posX, rawCam.posY, rawCam.posZ),
    rotation: new THREE.Euler().setFromQuaternion(
      new THREE.Quaternion(
        rawCam.angleX,
        rawCam.angleY,
        rawCam.angleZ,
        rawCam.angleW,
      ),
      "YXZ",
    ),
    fov: rawCam.fov,
    // mock aspect for now
    aspectWidth: 4,
    aspectHeight: 3,
    isHidingArrows: rawCam.isHidingArrows,
    isHidingWheels: rawCam.isHidingWheels,
    isLockingPosition: rawCam.isLockingPosition,
    isLockingRotation: rawCam.isLockingRotation,
    isHidingFrustum: rawCam.isHidingFrustum,
  };
}

function transformCamsData(
  modelWithCamsResp: ModelWithCamsResp,
): Record<string, ICamera> {
  return Object.fromEntries(
    Object.entries(modelWithCamsResp.data.cameras).map(([camId, rawCam]) => {
      const cam = transformProtoEventToCamera(rawCam);
      return [camId, cam];
    }),
  );
}

export function createBaseSceneStates(
  websocket: UseWebSocketReturn<unknown> | undefined,
  modelWithCamsResp: ModelWithCamsResp,
) {
  const tresContext = ref<TresContext | null>(null);

  const draggableObjects: Set<Obj3DWithUserData> = new Set();

  const isDraggingObject: Ref<boolean> = ref(false);

  const transformingInfo: Ref<
    | {
        position: THREE.Vector3;
        rotation: THREE.Euler;
        fov: number;
      }
    | undefined
  > = ref(undefined);

  const currentCamId: Ref<string | null> = ref(null);

  const spectatorCameraPosition: Reactive<THREE.Vector3> = reactive(
    new THREE.Vector3(0, 1, 0),
  );

  const spectatorCameraRotation: Reactive<THREE.Euler> = reactive(
    new THREE.Euler(0, 0, 0, "YXZ"),
  );

  const spectatorCameraFov: Ref<number> = ref(75);

  const tresCanvasParent: Ref<HTMLDivElement | null> = ref(null);

  const modelInfo = modelWithCamsResp;

  const camsData = transformCamsData(modelWithCamsResp);

  const cameras = reactive<Record<string, ICamera>>(camsData!);

  const spectatorCam: Reactive<ICamera> = reactive({
    ...structuredClone(cameraDefault),
    fov: spectatorCameraFov,
    position: spectatorCameraPosition,
    rotation: spectatorCameraRotation,
    aspectWidth: 0,
    aspectHeight: 1,
  });

  const currentCam = computed<ICamera>(() => {
    return currentCamId.value == null
      ? spectatorCam
      : cameras![currentCamId.value]!;
  });

  const markedForCheck = reactive(new Set<string>());

  const screenSize = reactive({
    width: undefined as number | undefined,
    height: undefined as number | undefined,
  });

  const aspectMarginType = ref<"horizontal" | "vertical">("horizontal");

  const aspectMargin = reactive({
    width: "0",
    height: "0",
  });

  const sceneStates = {
    tresContext,
    draggableObjects,
    isDraggingObject,
    currentCamId,
    currentCam,
    transformingInfo,
    spectatorCameraPosition,
    spectatorCameraRotation,
    spectatorCameraFov,
    tresCanvasParent,
    websocket,
    cameras,
    error: null,
    markedForCheck,
    modelInfo,
    screenSize,
    aspectMargin,
    aspectMarginType,
  } as const;

  // websocket.ws.value!.onclose = (_closeEvent: CloseEvent) => {
  //   sceneStates.websocket = useWebSocket(websocketUrl);
  // };

  return sceneStates;
}

export function createSceneStatesWithHelper(
  sceneStates: Awaited<BaseSceneStates>,
  workspace: string | null,
) {
  const aspectRatioManagement = useAspectRatioManagement(sceneStates);

  onMounted(() => {
    useAutosave(sceneStates, workspace);
  });

  const sceneStatesWithCam = {
    ...sceneStates,
    aspectRatioManagement: aspectRatioManagement,
    cameraManagement: useCameraManagement(sceneStates),
    spectatorPosition: useSpectatorPosition(sceneStates),
    spectatorRotation: useSpectatorRotation(sceneStates),
  };
  return sceneStatesWithCam;
}
