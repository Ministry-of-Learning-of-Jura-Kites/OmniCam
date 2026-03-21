import type { TresContext } from "@tresjs/core";
import type { Reactive } from "vue";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import type { SceneStates as BaseSceneStates } from "~/types/scene-states";
import { Quaternion, Euler, Vector3 } from "three";
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
  width: number;
  height: number;
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

  const localVersion = ref(modelInfo.data.version);
  const lastSyncedVersion = ref(modelInfo.data.version);

  const calibration = reactive({
    scale: modelWithCamsResp.data.scaleFactor ?? 1.0,
    heightOffset: modelWithCamsResp.data.modelHeight ?? 0.0,
    dirty: false,
  });
  const selectionMode = ref<"none" | "coverage-area">("none");
  const selectedCoverageFaces = ref<CoverageFace[]>([]);

  const setSelectionMode = (mode: "none" | "coverage-area") => {
    selectionMode.value = mode;
  };

  const isAllCoverageHidden = ref(false);

  const toggleAllCoverageHidden = () => {
    isAllCoverageHidden.value = !isAllCoverageHidden.value;
    tresContext.value?.invalidate?.();
  };

  const setAllCoverageHidden = (hidden: boolean) => {
    isAllCoverageHidden.value = hidden;
    tresContext.value?.invalidate?.();
  };

  const toggleCoverageFaceHidden = (faceId: string) => {
    const idx = selectedCoverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const next = [...selectedCoverageFaces.value];
    next[idx] = {
      ...next[idx]!,
      hidden: !next[idx]!.hidden,
    };

    selectedCoverageFaces.value = next;
    tresContext.value?.invalidate?.();
  };

  const setCoverageFaceHidden = (faceId: string, hidden: boolean) => {
    const idx = selectedCoverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const next = [...selectedCoverageFaces.value];
    next[idx] = {
      ...next[idx]!,
      hidden,
    };

    selectedCoverageFaces.value = next;
    tresContext.value?.invalidate?.();
  };

  const addCoverageFace = (face: CoverageFace) => {
    selectedCoverageFaces.value.push({
      ...face,
      color: face.color ?? "#22ff88",
      hidden: face.hidden ?? false,
    });
    tresContext.value?.invalidate?.();
  };

  const removeCoverageFace = (faceId: string) => {
    selectedCoverageFaces.value = selectedCoverageFaces.value.filter(
      (f) => f.id !== faceId,
    );
    tresContext.value?.invalidate?.();
  };

  const updateCoverageFaceColor = (faceId: string, color: string) => {
    const idx = selectedCoverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const next = [...selectedCoverageFaces.value];
    next[idx] = {
      ...next[idx]!,
      color,
    };

    selectedCoverageFaces.value = next;
    tresContext.value?.invalidate?.();
  };

  const clearCoverageFaces = () => {
    selectedCoverageFaces.value.splice(0, selectedCoverageFaces.value.length);
    tresContext.value?.invalidate?.();
  };

  const addOptimizedCamera = (cam: OptimizedCamera) => {
    cameras[cam.id] = {
      ...cameraDefault,
      name: cam.name,
      position: new Vector3(...cam.position),
      rotation: new Euler(
        cam.rotation[0],
        cam.rotation[1],
        cam.rotation[2],
        "YXZ",
      ),
      fov: cam.fov,
      aspectWidth: 1920,
      aspectHeight: 1080,
      isHidingArrows: false,
      isHidingWheels: false,
      isLockingPosition: false,
      isLockingRotation: false,
      isHidingFrustum: false,
      frustumColor: { r: 0, g: 255, b: 100, a: 1 },
      frustumLength: 10,
    };
  };

  const updateCoverageFaceCorner = (
    faceId: string,
    cornerIndex: number,
    point: [number, number, number],
  ) => {
    const idx = selectedCoverageFaces.value.findIndex((f) => f.id === faceId);
    if (idx < 0) return;

    const face = selectedCoverageFaces.value[idx]!;
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

    const dist = (a: [number, number, number], b: [number, number, number]) =>
      Math.hypot(a[0] - b[0], a[1] - b[1], a[2] - b[2]);

    const next = [...selectedCoverageFaces.value];
    next[idx] = {
      ...face,
      points: newPoints,
      center,
      width: dist(newPoints[0]!, newPoints[1]!),
      height: dist(newPoints[1]!, newPoints[2]!),
    };

    selectedCoverageFaces.value = next;
  };

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
    // Predictive Camera Placement
    selectionMode,
    selectedCoverageFaces,
    setSelectionMode,
    clearCoverageFaces,
    addOptimizedCamera,
    addCoverageFace,
    updateCoverageFaceCorner,
    removeCoverageFace,
    updateCoverageFaceColor,
    isAllCoverageHidden,
    toggleAllCoverageHidden,
    setAllCoverageHidden,
    toggleCoverageFaceHidden,
    setCoverageFaceHidden,
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
