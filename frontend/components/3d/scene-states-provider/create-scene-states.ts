import type { TresContext } from "@tresjs/core";
import type { Reactive } from "vue";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import type { SceneStates as BaseSceneStates } from "~/types/scene-states";
import {
  Quaternion,
  Euler,
  Vector3,
  type PerspectiveCamera,
  type CubeCamera,
} from "three";
import { cameraDefault, type ICamera } from "~/types/camera";
import { useCameraManagement } from "../scene-3d/use-camera-management";
import { useSpectatorRotation } from "../scene-3d/use-spectator-rotation";
import { useSpectatorPosition } from "../scene-3d/use-spectator-position";
import type { UseWebSocketReturn } from "@vueuse/core";
import type { Camera } from "~/messages/protobufs/autosave_event";
import { useAspectRatio as useAspectRatioManagement } from "../scene-3d/use-aspect-ratio";
import { useAutosave } from "../scene-3d/use-autosave";

export interface CoverageFace {
  id: string;
  points: [number, number, number][];
  center: [number, number, number];
  normal?: [number, number, number];
  planeY?: number;
  color?: string;
  hidden?: boolean;
}
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
    fileExtension: string;
    imagePath: string;
    imageExtension: string;
    cameras: Record<string, Camera>;
    scaleFactor?: number;
    modelHeight?: number;
  };
}

export interface OptimizedCamera {
  id: string;
  name: string;
  position: [number, number, number];
  rotation: [number, number, number];
  fov: number;
}

export function transformProtoEventToCamera(rawCam: Camera): ICamera {
  return {
    name: rawCam.name,
    position: new Vector3(rawCam.posX, rawCam.posY, rawCam.posZ),
    rotation: new Euler().setFromQuaternion(
      new Quaternion(
        rawCam.angleX,
        rawCam.angleY,
        rawCam.angleZ,
        rawCam.angleW,
      ),
      "YXZ",
    ),
    fov: rawCam.fov,
    // mock aspect for now
    // aspectWidth: rawCam.aspectWidth,
    // aspectHeight: rawCam.aspectHeight,
    widthRes: rawCam.widthRes,
    heightRes: rawCam.heightRes,
    isHidingArrows: rawCam.isHidingArrows,
    isHidingWheels: rawCam.isHidingWheels,
    distortion: rawCam.distortion ?? structuredClone(cameraDefault.distortion),
    isLockingPosition: rawCam.isLockingPosition,
    isLockingRotation: rawCam.isLockingRotation,
    isHidingFrustum: rawCam.isHidingFrustum,
    frustumColor: rawCam.frustumColor!,
    frustumLength: rawCam.frustumLength,
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
        position: Vector3;
        rotation: Euler;
        fov: number;
      }
    | undefined
  > = ref(undefined);

  const currentCamId: Ref<string | null> = ref(null);

  const spectatorCameraPosition: Reactive<Vector3> = reactive(
    new Vector3(0, 1, 0),
  );

  const spectatorCameraRotation: Reactive<Euler> = reactive(
    new Euler(0, 0, 0, "YXZ"),
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
    widthRes: 0,
    heightRes: 1,
    frustumColor: { r: 0, g: 0, b: 0, a: 0 },
    frustumLength: 0,
  });

  const currentCam = computed<ICamera>(() => {
    return currentCamId.value == null
      ? spectatorCam
      : cameras![currentCamId.value]!;
  });

  const markedForCheck = ref<boolean>(false);

  const screenSize = reactive({
    width: undefined as number | undefined,
    height: undefined as number | undefined,
  });

  const aspectMarginType = ref<"horizontal" | "vertical">("horizontal");

  const aspectMargin = reactive({
    width: "0",
    height: "0",
  });

  const localVersion = ref(modelInfo.data.version);
  const lastSyncedVersion = ref(modelInfo.data.version);

  const calibration = reactive({
    scale: modelWithCamsResp.data.scaleFactor ?? 1.0,
    heightOffset: modelWithCamsResp.data.modelHeight ?? 0.0,
    dirty: false,
  });
  const currentIsFisheye = computed(() => {
    if (currentCamId.value != null) {
      return cameras[currentCamId.value]!.distortion.isFisheye;
    }
    return false;
  });

  const currentFov = computed(() => {
    return currentCam.value.fov;
  });

  const currentDistEnabled = computed(() => {
    if (currentCamId.value == undefined) {
      return false;
    }
    return currentCam.value.distortion.enabled;
  });
  const aspectRatio = computed<number>(() => {
    if (screenSize.width == null || screenSize.height == null) {
      return 1;
    }
    return screenSize.width! / screenSize.height!;
  });

  const perspectiveCamera = ref<PerspectiveCamera | null>(null);
  const cubeCamera = ref<CubeCamera | null>(null);

  const coverageMode = ref<"none" | "coverage-area">("none");
  const coverageFaces = ref<CoverageFace[]>([]);
  const coverageAllHidden = ref(false);

  const setCoverageMode = (mode: "none" | "coverage-area") => {
    coverageMode.value = mode;
  };

  const toggleAllCoverageHidden = () => {
    coverageAllHidden.value = !coverageAllHidden.value;
  };

  const setAllCoverageHidden = (hidden: boolean) => {
    coverageAllHidden.value = hidden;
  };

  const toggleCoverageFaceHidden = (faceId: string) => {
    const idx = coverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const next = [...coverageFaces.value];
    next[idx] = {
      ...next[idx]!,
      hidden: !next[idx]!.hidden,
    };

    coverageFaces.value = next;
  };

  const setCoverageFaceHidden = (faceId: string, hidden: boolean) => {
    const idx = coverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const next = [...coverageFaces.value];
    next[idx] = {
      ...next[idx]!,
      hidden,
    };

    coverageFaces.value = next;
  };

  const addCoverageFace = (face: CoverageFace) => {
    coverageFaces.value.push({
      ...face,
      color: face.color ?? "#22ff88",
      hidden: face.hidden ?? false,
    });
  };

  const removeCoverageFace = (faceId: string) => {
    coverageFaces.value = coverageFaces.value.filter((f) => f.id !== faceId);
  };

  const updateCoverageFaceColor = (faceId: string, color: string) => {
    const idx = coverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const next = [...coverageFaces.value];
    next[idx] = {
      ...next[idx]!,
      color,
    };

    coverageFaces.value = next;
  };

  const clearCoverageFaces = () => {
    coverageFaces.value = [];
  };

  const updateCoverageFaceCorner = (
    faceId: string,
    cornerIndex: number,
    point: [number, number, number],
  ) => {
    const idx = coverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const face = coverageFaces.value[idx]!;
    const newPoints = face.points.map((p, i) =>
      i === cornerIndex ? point : p,
    ) as [number, number, number][];

    const center: [number, number, number] = [
      (newPoints[0]![0] +
        newPoints[1]![0] +
        newPoints[2]![0] +
        newPoints[3]![0]) /
        4,
      (newPoints[0]![1] +
        newPoints[1]![1] +
        newPoints[2]![1] +
        newPoints[3]![1]) /
        4,
      (newPoints[0]![2] +
        newPoints[1]![2] +
        newPoints[2]![2] +
        newPoints[3]![2]) /
        4,
    ];

    const next = [...coverageFaces.value];
    next[idx] = {
      ...face,
      points: newPoints,
      center,
    };

    coverageFaces.value = next;
  };

  const facesManagement = {
    mode: coverageMode,
    faces: coverageFaces,
    isAllHidden: coverageAllHidden,
    setMode: setCoverageMode,
    clear: clearCoverageFaces,
    add: addCoverageFace,
    updateCorner: updateCoverageFaceCorner,
    remove: removeCoverageFace,
    updateColor: updateCoverageFaceColor,
    toggleAllHidden: toggleAllCoverageHidden,
    setAllHidden: setAllCoverageHidden,
    toggleFaceHidden: toggleCoverageFaceHidden,
    setFaceHidden: setCoverageFaceHidden,
  } as const;

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
    localVersion,
    lastSyncedVersion,
    calibration,
    currentDistEnabled,
    currentFov,
    currentIsFisheye,
    aspectRatio,
    perspectiveCamera,
    cubeCamera,
    facesManagement,
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
  useAutosave(sceneStates, workspace);

  onMounted(() => {
    useAutosave(sceneStates, workspace);
    watch(
      () => [sceneStates.transformingInfo, sceneStates.currentCam],
      ([transform, cam]) => {
        const newFov = transform?.value?.fov ?? cam?.value?.fov;
        const actualCamera = sceneStates.tresContext.value?.camera.activeCamera;
        if (actualCamera && newFov !== undefined) {
          (actualCamera as PerspectiveCamera).fov = newFov;
          actualCamera.updateProjectionMatrix();
        }
      },
      { deep: true },
    );
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
