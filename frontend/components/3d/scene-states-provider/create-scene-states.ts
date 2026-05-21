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
import { useAspectRatio as useAspectRatioManagement } from "../scene-3d/use-aspect-ratio";
import type { GLTF } from "three-stdlib";
import type { QuadrilateralPoints as QuadrilateralPoints } from "~/types/trapezoid";
import { useOptimize } from "../scene-3d/use-optimize";
import type { Camera } from "~/messages/protobufs/camera";
import type { CoverageFace } from "~/messages/protobufs/optimization";
import type { ProtoVector3 } from "~/messages/protobufs/vector";
import { useAutosave } from "~/components/3d/scene-3d/use-autosave";
import type { WorkspaceEventResponse } from "~/messages/protobufs/workspace_event";
import { useLivestream } from "../scene-3d/use-livestream";
import { arrayPointsToNormal } from "~/utils/face-helper/get-avg-normal";
import { v4 as uuidv4 } from "uuid";

export interface ProcessedCoverageFace {
  name: string;
  points: QuadrilateralPoints;
  color: string | undefined;
  hidden: boolean;
  normal: Vector3;
  // Derived field
  center?: [number, number, number];
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
    targetTrapezoids?: Record<string, CoverageFace>;
    // Realife meter / virtual meter
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

function protoVecToNumbers(vec: ProtoVector3): [number, number, number] {
  return [vec.x, vec.y, vec.z];
}

function transformProtoEventToTrapezoid(
  rawTrapezoid: CoverageFace,
): ProcessedCoverageFace {
  const points = rawTrapezoid.points.map(
    protoVecToNumbers,
  ) as QuadrilateralPoints;
  return {
    name: rawTrapezoid.name,
    points: points,
    color: rawTrapezoid.color,
    hidden: rawTrapezoid.hidden,
    normal: arrayPointsToNormal(points),
  };
}

function transformFacesData(
  modelWithCamsResp: ModelWithCamsResp,
): Record<string, ProcessedCoverageFace> {
  if (modelWithCamsResp.data.targetTrapezoids == undefined) {
    return {};
  }
  return Object.fromEntries(
    Object.entries(modelWithCamsResp.data.targetTrapezoids).map(
      ([id, trapezoid]) => {
        const transTrapezoid = transformProtoEventToTrapezoid(trapezoid);
        return [id, transTrapezoid];
      },
    ),
  );
}

export function createBaseSceneStates(
  autosaveWebsocket: UseWebSocketReturn<unknown> | undefined,
  livestreamWebsocket: UseWebSocketReturn<unknown> | undefined,
  modelWithCamsResp: ModelWithCamsResp,
  workspace: string | undefined,
) {
  const tresContext = ref<TresContext | null>(null);

  const draggableObjects: Set<Obj3DWithUserData> = new Set();

  const clickableObjects: Set<Obj3DWithUserData> = new Set();

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

  const markedFacesForCheck = ref(false);

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

  const selectionMode = ref<"none" | "coverage-area" | "measurement">("none");
  const initialCoverageFaces = transformFacesData(modelWithCamsResp);
  const coverageFaces =
    reactive<Record<string, ProcessedCoverageFace>>(initialCoverageFaces);
  const coverageAllHidden = ref(false);

  const setCoverageMode = (mode: "none" | "coverage-area" | "measurement") => {
    selectionMode.value = mode;
  };

  const toggleAllCoverageHidden = () => {
    coverageAllHidden.value = !coverageAllHidden.value;
  };

  const setAllCoverageHidden = (hidden: boolean) => {
    coverageAllHidden.value = hidden;
  };

  const toggleCoverageFaceHidden = (faceId: string) => {
    const found = coverageFaces[faceId];
    if (found == undefined) return;

    found.hidden = !found.hidden;
  };

  const setCoverageFaceHidden = (faceId: string, hidden: boolean) => {
    const found = coverageFaces[faceId];
    if (found == undefined) return;

    found.hidden = hidden;
  };

  const addCoverageFace = (id: string, face: ProcessedCoverageFace) => {
    coverageFaces[id] = {
      ...face,
      color: face.color ?? "#22ff88",
      hidden: face.hidden ?? false,
    };
  };

  const removeCoverageFace = (faceId: string) => {
    delete coverageFaces[faceId];
  };

  const updateCoverageFaceColor = (faceId: string, color: string) => {
    const found = coverageFaces[faceId];
    if (found == undefined) return;

    found.color = color;
  };

  const clearCoverageFaces = () => {
    // Remove this function later
    for (const face of Object.keys(coverageFaces)) {
      delete coverageFaces[face];
    }
  };

  const updateCoverageFaceCorner = (
    faceId: string,
    cornerIndex: number,
    point: [number, number, number],
  ) => {
    const found = coverageFaces[faceId];
    if (found == undefined) return;

    const newPoints = found.points.map((p, i) =>
      i === cornerIndex ? point : p,
    ) as QuadrilateralPoints;

    // Calculate Center
    const center: [number, number, number] = [
      (newPoints[0][0] + newPoints[1][0] + newPoints[2][0] + newPoints[3][0]) /
        4,
      (newPoints[0][1] + newPoints[1][1] + newPoints[2][1] + newPoints[3][1]) /
        4,
      (newPoints[0][2] + newPoints[1][2] + newPoints[2][2] + newPoints[3][2]) /
        4,
    ];

    // We use the same CCW-reliable logic from the build function
    const p0 = new Vector3(...newPoints[0]);
    const p1 = new Vector3(...newPoints[1]);
    const p2 = new Vector3(...newPoints[2]);

    const e1 = new Vector3().subVectors(p1, p0);
    const e2 = new Vector3().subVectors(p2, p0);
    const newNormal = new Vector3().crossVectors(e1, e2).normalize();

    // 3. Update the object
    found.points = newPoints;
    found.center = center;
    found.normal = newNormal; // Now the Gizmo knows the new orientation!
  };

  const facesManagement = {
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

  const modelRef = ref<GLTF | null>(null);

  const measurement = reactive({
    draftStartPoint: null as Vector3 | null,

    lines: [] as {
      id: string;
      label: string;
      start: Vector3;
      end: Vector3;
      virtualDistance: number;
      realDistance: number;
    }[],

    addLine(start: Vector3, end: Vector3) {
      const virtualdistance = start.distanceTo(end);
      const realDistance = virtualdistance * calibration.scale;

      const index = this.lines.length + 1;

      this.lines.push({
        id: uuidv4(),
        label: `Measurement ${index} - ${realDistance.toFixed(2)}m`,
        start: start.clone(),
        end: end.clone(),
        virtualDistance: virtualdistance,
        realDistance: realDistance,
      });
    },
    updateEndpoint(lineId: string, pointIndex: 0 | 1, newPos: Vector3) {
      const line = this.lines.find((l) => l.id === lineId);
      if (!line) return;

      if (pointIndex === 0) line.start = newPos.clone();
      else line.end = newPos.clone();

      const virtualDistance = line.start.distanceTo(line.end);
      const realDistance = virtualDistance * calibration.scale;
      line.virtualDistance = virtualDistance;
      line.realDistance = realDistance;

      const prefix = line.label.split(" - ")[0];
      line.label = `${prefix} - ${realDistance.toFixed(2)}m`;
    },

    resetDraft() {
      this.draftStartPoint = null;
    },
  });

  const miniScene = reactive({
    visible: false,
    targetCameraId: null as string | null,
  });

  const sceneStates = {
    tresContext,
    modelRef,
    selectionMode,
    draggableObjects,
    clickableObjects,
    isDraggingObject,
    currentCamId,
    currentCam,
    transformingInfo,
    spectatorCam,
    spectatorCameraPosition,
    spectatorCameraRotation,
    spectatorCameraFov,
    tresCanvasParent,
    websocket: autosaveWebsocket,
    livestreamWebsocket,
    cameras,
    error: null,
    markedForCheck,
    markedFacesForCheck,
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
    measurement,
    miniScene,
  } as const;

  // websocket.ws.value!.onclose = (_closeEvent: CloseEvent) => {
  //   sceneStates.websocket = useWebSocket(websocketUrl);
  // };

  if (workspace == "me") {
    watch(
      cameras,
      () => {
        markedForCheck.value = true;
      },
      { deep: true },
    );
  }

  return sceneStates;
}

export function createSceneStatesWithHelper(
  sceneStates: Awaited<BaseSceneStates>,
  workspace: string | null,
) {
  const livestream = useLivestream(sceneStates, workspace);
  const aspectRatioManagement = useAspectRatioManagement(sceneStates);

  const optimization = useOptimize(sceneStates, workspace);

  function handle(resp: WorkspaceEventResponse) {
    if (resp.autosave) {
      sceneStates.lastSyncedVersion.value = resp.autosave.lastUpdatedVersion;
    } else {
      const submitStatus = optimization!.submitStatus!;
      if (!resp.optimize) {
        return;
      }

      if (resp.optimize.successResp) {
        console.log(resp.optimize.successResp.cameras);
        for (const cam of resp.optimize.successResp.cameras) {
          optimization!.candidateCameras[cam.id] =
            transformProtoEventToCamera(cam);
        }
      }
      submitStatus.value = "idle";
    }
  }

  useAutosave(sceneStates, workspace, handle);

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

  const sceneStatesWithCam = {
    ...sceneStates,
    aspectRatioManagement: aspectRatioManagement,
    cameraManagement: useCameraManagement(sceneStates),
    spectatorPosition: useSpectatorPosition(sceneStates, workspace),
    spectatorRotation: useSpectatorRotation(sceneStates, workspace),
    optimization,
    livestream,
  };
  return sceneStatesWithCam;
}
