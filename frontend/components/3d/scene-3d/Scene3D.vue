<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import AdjustableInput from "../../adjustable-input/AdjustableInput.vue";
import { CAMERA_UTILS_LAYER, SPECTATOR_ADJ_INPUT_SENTIVITY } from "~/constants";
import CameraObject from "../camera-object/CameraObject.vue";
import {
  type PerspectiveCamera,
  Raycaster,
  Vector2,
  DoubleSide,
  Vector3,
  Matrix3,
} from "three";
import { MAP_KEY, PANEL_KEY, SCENE_STATES_KEY } from "@/constants/state-keys";
import Stats from "stats.js";
import LazyMinimap from "@/components/3d/minimap/Minimap.vue";
import { useCameraUpdate } from "./use-camera-update";
import type { IUserData } from "~/types/obj-3d-user-data";
import ModelLoader from "../model-loader/ModelLoader.vue";
import CalibrationGrid from "../calibration/CalibrationGrid.vue";
import { usePromptUnsaved } from "./use-prompt-unsaved";
import FrustumOverlay from "@/components/3d/camera-frustum/FrustumOverlay.vue";
import CoverageAreaMesh from "../coverage-area-mesh/CoverageAreaMesh.vue";
import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";
import CoverageCornerGizmo from "../coverage-area-mesh/CoverageCornerGizmo.vue";
import { averageVector } from "~/utils/face-helper/avg-vec";
import { v4 as uuidv4 } from "uuid";
import { get3dModelPathClient } from "~/composables/api/use-fetch-model-api";
import AxisGizmo from "../axis-gizmo/axis-gizmo.vue";
import LineMeasurement from "../line-measurement/line-measurement.vue";
import type {
  QuadrilateralPoints,
  QuadrilateralVectors,
} from "~/types/trapezoid";
import CameraDirection from "../camera-direction/CameraDirection.vue";
import MiniCameraScene from "../mini-camera-scene/MiniCameraScene.vue";
import FailDialog from "~/components/dialog/FailDialog.vue";
// import { watchDebounced } from "@vueuse/core";

const { isPanelOpen, currentPanel, camPanelInfo } = inject(PANEL_KEY)!;
const { selectedCamId } = camPanelInfo;

const router = useRouter();
const route = useRoute();

const selectedFaces = computed(() =>
  Object.entries(
    sceneStates.value!.facesManagement.faces ?? ({} as ProcessedCoverageFace),
  ).filter(
    ([_id, face]) =>
      !sceneStates.value!.facesManagement.isAllHidden.value && !face.hidden,
  ),
);
type Point3 = [number, number, number];
const props = withDefaults(
  defineProps<{
    projectId: string;
    modelId: string;
    workspace?: string | null;
  }>(),
  {
    workspace: null,
  },
);

const isFailedDialogOpen = ref<boolean>(false);
const failedMessage = ref<string>("");
// line measurement
const lineMeasurement = ref<InstanceType<typeof LineMeasurement> | null>(null);

const config = useRuntimeConfig();
const sceneStates = inject(SCENE_STATES_KEY)!;
const modelResp = computed(() => {
  return sceneStates.value!.modelInfo.data;
});

const modelPath = get3dModelPathClient(
  props.projectId,
  props.modelId,
  modelResp.value.fileExtension,
);

const perspectiveCamera = ref<PerspectiveCamera | null>(null);
const canvas: Ref<InstanceType<typeof TresCanvas> | null> = ref(null);

const { isMapOpen } = inject(MAP_KEY)!;

const COVERAGE_Y_OFFSET = 0.01;

const draftCoveragePoints = ref<Vector3[]>([]);

const aspect = computed(() => {
  const width = sceneStates.value!.currentCam.value.widthRes;
  if (width == 0) return undefined;
  const height = sceneStates.value!.currentCam.value.heightRes || 1;
  return width / height;
});

const previewPoints = computed<Point3[]>(() => {
  if (sceneStates.value!.selectionMode.value !== "coverage-area") return [];

  return buildDraftCoveragePreview(draftCoveragePoints.value);
});

const isPreviewing = computed(() => previewPoints.value.length === 4);

const draftPointMarkers = computed<Point3[]>(() => {
  if (sceneStates.value!.selectionMode.value !== "coverage-area") return [];

  return draftCoveragePoints.value.map((p) => [p.x, p.y, p.z] as Point3);
});

const selectedCam = computed(() => {
  if (selectedCamId.value == null) {
    return null;
  }
  return sceneStates.value!.cameras[selectedCamId.value];
});

const lineColors = reactive<Record<string, string>>({});

const MEASUREMENT_PALETTE = [
  "#f97316", // orange
  "#a855f7", // purple
  "#06b6d4", // cyan
  "#f59e0b", // amber
  "#ec4899", // pink
  "#84cc16", // lime
  "#14b8a6", // teal
  "#fb923c", // light orange
];

function getNextMeasurementColor(): string {
  const used = Object.keys(lineColors).length;
  return MEASUREMENT_PALETTE[used % MEASUREMENT_PALETTE.length]!;
}

usePromptUnsaved(sceneStates.value!);

useCameraUpdate(sceneStates.value!);

// ── Raycasting & Input Events (Omitted same logic for brevity) ───────
const raycaster = new Raycaster();
raycaster.layers.enable(CAMERA_UTILS_LAYER);

const mouse = new Vector2();

function clearDraftCoverageSelection() {
  draftCoveragePoints.value = [];
  // sceneStates.value!.tresContext.value?.invalidate?.();
}

function buildDraftCoveragePreview(points: Vector3[]): Point3[] {
  const count = points.length;
  if (count < 1) return [];

  if (count < 3) {
    return points.map((p) => [p.x, p.y, p.z] as Point3);
  }

  const sorted = sortPointsConvex(points);
  const p0 = sorted[0]!;
  const p1 = sorted[1]!;
  const p2 = sorted[2]!;
  // 3 points EXACTLY: Return a triangle (P0 -> P1 -> P2)
  if (count === 3) {
    return points.map((p) => [p.x, p.y, p.z] as Point3);
  }

  if (count === 4) {
    const p3Raw = sorted[3]!;

    const dir = new Vector3().subVectors(p2, p1).normalize();
    const v = new Vector3().subVectors(p3Raw, p0);
    const dot = v.dot(dir);
    const p3 = p0.clone().add(dir.multiplyScalar(dot));

    return [p0, p1, p2, p3].map((p) => [p.x, p.y, p.z] as Point3);
  }

  return sorted.map((p) => [p.x, p.y, p.z] as Point3);
}

const defaultCoverageFace: ProcessedCoverageFace = {
  name: "",
  points: [
    [0, 0, 0],
    [0, 0, 0],
    [0, 0, 0],
    [0, 0, 0],
  ],
  normal: new Vector3(0, 0, 1),
  color: "#22ff88",
  hidden: false,
};

function sortPointsConvex(points: Vector3[]): Vector3[] {
  const center = new Vector3();
  points.forEach((p) => center.add(p));
  center.divideScalar(points.length);

  const e1 = new Vector3().subVectors(points[1]!, points[0]!).normalize();
  const e2 = new Vector3(0, 1, 0);
  const normal = new Vector3().crossVectors(e1, e2).normalize();
  const yAxis = new Vector3().crossVectors(normal, e1).normalize();

  return [...points].sort((a, b) => {
    const da = new Vector3().subVectors(a, center);
    const db = new Vector3().subVectors(b, center);
    const angleA = Math.atan2(da.dot(yAxis), da.dot(e1));
    const angleB = Math.atan2(db.dot(yAxis), db.dot(e1));
    return angleA - angleB;
  });
}

function buildCoverageFaceFromPickedPoints(
  points: QuadrilateralVectors,
): ProcessedCoverageFace | null {
  if (points.length !== 4) return null;
  const sorted = sortPointsConvex(points);

  const p0 = sorted[0]!;
  const p1 = sorted[1]!;
  const p2 = sorted[2]!;
  const p3Raw = sorted[3]!;

  // 1. Force Trapezoid/Parallelism
  const dir = new Vector3().subVectors(p2, p1).normalize();
  const v = new Vector3().subVectors(p3Raw, p0);
  const p3 = p0.clone().add(dir.multiplyScalar(v.dot(dir)));

  const finalPoints = [p0, p1, p2, p3];
  const centerV = averageVector(finalPoints);

  // 2. Initial Normal Calculation (Vector3)
  const e1 = new Vector3().subVectors(p1, p0);
  const e2 = new Vector3().subVectors(p2, p0);
  const finalNormal = new Vector3().crossVectors(e1, e2).normalize();

  // 4. Final Validation
  const width = p0.distanceTo(p1);
  const height = p1.distanceTo(p2);
  if (width < 0.05 || height < 0.05) {
    console.error("Face too small, rejecting.");
    return null;
  }

  return {
    ...defaultCoverageFace,
    points: finalPoints.map(threeVector3ToNumbers) as QuadrilateralPoints,
    normal: finalNormal, // Maintained as Vector3
    center: [centerV.x, centerV.y, centerV.z],
  };
}

function handleMeasurementPointer(event: PointerEvent) {
  const measurement = sceneStates.value!.measurement!;

  const isModifierPressed = event.ctrlKey || event.altKey;

  if (event.type !== "pointerdown" || !isModifierPressed) {
    return false;
  }

  const hit = getSurfaceHit(event);

  if (!hit) {
    return false;
  }

  const point = hit.point.clone();

  // first click
  if (!measurement.draftStartPoint) {
    measurement.draftStartPoint = point;
    return true;
  }

  measurement.addLine(measurement.draftStartPoint, point);

  measurement.resetDraft();

  return true;
}

function worldToScreen(position: Vector3) {
  const camera = perspectiveCamera.value;
  if (
    !camera ||
    !sceneStates!.value?.screenSize.width ||
    !sceneStates!.value?.screenSize.height
  ) {
    return null;
  }

  const projected = position.clone().project(camera);

  // Point is behind the camera
  if (projected.z > 1) return null;

  return {
    x: (projected.x + 1) * 0.5 * sceneStates!.value.screenSize.width,
    y: (-projected.y + 1) * 0.5 * sceneStates!.value.screenSize.height,
  };
}

function getMeasurementLabelStyle(line: { start: Vector3; end: Vector3 }) {
  const worldPos = line.start.clone().add(new Vector3(0, 0.08, 0));
  const screen = worldToScreen(worldPos);

  if (!screen) return { display: "none" };

  return {
    display: "block",
    left: `${screen.x}px`,
    top: `${screen.y}px`,
    transform: "translate(-50%, -100%)",
  };
}

const labelStyles = ref<
  Record<string, ReturnType<typeof getMeasurementLabelStyle>>
>({});

function updateLabelStyles() {
  if (sceneStates.value?.measurement?.lines?.length) {
    const updated: typeof labelStyles.value = {};
    for (const line of sceneStates.value.measurement.lines) {
      updated[line.id] = getMeasurementLabelStyle(line);
    }
    labelStyles.value = updated;
  }
}

let labelRafId: number | null = null;

function labelLoop() {
  updateLabelStyles();
  labelRafId = requestAnimationFrame(labelLoop);
}

onMounted(() => {
  labelRafId = requestAnimationFrame(labelLoop);
});

onUnmounted(() => {
  if (labelRafId !== null) {
    cancelAnimationFrame(labelRafId);
    labelRafId = null;
  }
});

function handleCoverageAreaPointer(event: PointerEvent) {
  if (sceneStates.value!.selectionMode.value !== "coverage-area") return false;

  const isModifierPressed = event.ctrlKey || event.metaKey;

  if (event.type !== "pointerdown" || !isModifierPressed) {
    return false;
  }

  if (draftCoveragePoints.value.length >= 4) {
    clearDraftCoverageSelection();
  }

  const hit = getSurfaceHit(event);

  if (!hit) return false;

  const forward = new Vector3(0, 0, 1);
  forward.applyEuler(sceneStates.value!.spectatorCameraRotation!);

  draftCoveragePoints.value = [...draftCoveragePoints.value, hit.point.clone()];

  if (draftCoveragePoints.value.length === 4) {
    const face = buildCoverageFaceFromPickedPoints(
      draftCoveragePoints.value as QuadrilateralVectors,
    );

    if (face) {
      sceneStates.value!.facesManagement.add(uuidv4(), face);
    }

    requestAnimationFrame(() => {
      clearDraftCoverageSelection();
    });

    return true;
  }

  return true;
}
function getSurfaceHit(
  event: PointerEvent,
): { point: Vector3; normal: Vector3 } | null {
  const context = sceneStates.value!.tresContext.value;
  const camera = perspectiveCamera.value;

  if (!context?.renderer?.instance || !camera) return null;

  const canvas = context.renderer.instance.domElement;
  const rect = canvas.getBoundingClientRect();

  const x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  const y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

  mouse.set(x, y);
  raycaster.setFromCamera(mouse, camera);

  const hits = raycaster.intersectObjects(
    [sceneStates.value!.modelRef.value!.scene!],
    true,
  );
  const hit = hits[0];

  const forwardDirection = new Vector3();
  camera.getWorldDirection(forwardDirection);

  if (!hit || !hit.face) {
    const targetPoint = new Vector3();
    raycaster.ray.at(5, targetPoint);
    const normal = raycaster.ray.direction.clone().negate();
    return {
      point: targetPoint,
      normal: normal,
    };
  }
  const worldNormal = hit.face.normal
    .clone()
    .applyMatrix3(new Matrix3().getNormalMatrix(hit.object.matrixWorld))
    .normalize();

  return {
    point: hit.point.clone(),
    normal: worldNormal,
  };
}

function onCanvasKeydown(event: KeyboardEvent) {
  if (
    event.code == "Escape" &&
    sceneStates.value!.selectionMode.value === "coverage-area"
  ) {
    requestAnimationFrame(() => {
      clearDraftCoverageSelection();
    });
    return;
  }
  sceneStates.value!.spectatorPosition.onKeyDown(event);
}

function onCanvasPointer(event: PointerEvent) {
  if (!sceneStates.value!.tresContext.value || !perspectiveCamera.value) return;
  const ele = sceneStates.value!.tresContext.value.renderer.instance.domElement;
  const rect = ele.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height!) * 2 + 1;
  raycaster.setFromCamera(mouse, perspectiveCamera.value!);

  if (lineMeasurement.value?.onPointerEvent(event, raycaster)) return;

  if (sceneStates.value!.selectionMode.value === "coverage-area") {
    const handled = handleCoverageAreaPointer(event);
    if (handled) return;
  }
  if (currentPanel.value === "measurement") {
    const handled = handleMeasurementPointer(event);
    if (handled) return;
  }

  const objectsToSearch = [...sceneStates.value!.draggableObjects];
  if (event.type === "pointerdown" || event.type === "pointerup") {
    for (const obj of sceneStates.value!.clickableObjects) {
      objectsToSearch.push(obj);
    }
  }

  const intersects = raycaster.intersectObjects(objectsToSearch, false);
  if (intersects.length > 0) {
    const foundObj = intersects[0];
    const userData = foundObj?.object.userData as IUserData;
    userData.handleEvent.call(userData, event.type, event);
  } else if (
    event.type === "pointerdown" &&
    (sceneStates.value!.currentCamId.value == null || props.workspace == "me")
  ) {
    sceneStates.value!.spectatorRotation.onPointerDown(event);
  }
}

let stats: Stats | null = null;

onMounted(() => {
  const stopPersWatch = watch(
    perspectiveCamera,
    (camera) => {
      if (camera != undefined) {
        camera.layers.enable(CAMERA_UTILS_LAYER);
        sceneStates.value!.perspectiveCamera.value = camera;
        stopPersWatch();
      }
    },
    { immediate: true },
  );

  watch(
    () => canvas.value?.context,
    (context) => {
      if (!context) return;
      const renderer = context.renderer;
      if (stats != null) {
        renderer.loop.onBeforeLoop(() => {
          stats!.begin();
        });
        renderer.loop.onLoop(() => {
          stats!.end();
        });
      }

      sceneStates.value!.tresContext.value = context;
      renderer.instance.domElement.addEventListener(
        "pointerdown",
        onCanvasPointer,
      );
      renderer.instance.domElement.addEventListener(
        "pointermove",
        onCanvasPointer,
      );
      renderer.instance.domElement.addEventListener(
        "pointerup",
        onCanvasPointer,
      );
      renderer.instance.domElement.addEventListener("keydown", onCanvasKeydown);
      renderer.instance.domElement.addEventListener(
        "keyup",
        sceneStates.value!.spectatorPosition.onKeyUp,
      );
      renderer.instance.domElement.addEventListener("blur", (e: FocusEvent) => {
        sceneStates.value!.spectatorRotation.onBlur(e);
        sceneStates.value!.spectatorPosition.onBlur(e);
      });

      renderer.instance.domElement.addEventListener(
        "contextmenu",
        (event: Event) => {
          event.preventDefault();

          sceneStates.value!.spectatorRotation.onBlur(
            event as unknown as FocusEvent,
          );
          sceneStates.value!.spectatorPosition.onBlur(
            event as unknown as FocusEvent,
          );
        },
      );
    },
    { once: true },
  );
});

if (config.public.devMode) {
  onMounted(() => {
    stats = new Stats();
    stats.showPanel(0);
    if (document) {
      document.body.appendChild(stats.dom);
    }
  });
}
watch(
  () => sceneStates.value!.selectionMode.value,
  (mode) => {
    if (mode !== "coverage-area") {
      clearDraftCoverageSelection();
    }
  },
);

watch(
  () =>
    Object.keys(sceneStates.value!.facesManagement.faces.value ?? {}).length,
  (len) => {
    if (len === 0) {
      clearDraftCoverageSelection();
    }
  },
);

function selectCurrentCamShortcut() {
  const currentCamId = sceneStates.value!.currentCamId.value;
  if (currentCamId) {
    selectedCamId.value = sceneStates.value!.currentCamId.value;
    isPanelOpen.value = true;
    currentPanel.value = "camera";
  }
}
watch(
  () => sceneStates.value?.errorLivestreamMessage.value,
  (errorMessage) => {
    if (!errorMessage) return;

    isFailedDialogOpen.value = true;
    failedMessage.value = errorMessage;
  },
);

watch(
  () => sceneStates.value!.measurement.lines.length,
  () => {
    for (const line of sceneStates.value!.measurement.lines) {
      if (!lineColors[line.id]) {
        lineColors[line.id] = getNextMeasurementColor();
      }
    }
    // clean up removed lines
    for (const id of Object.keys(lineColors)) {
      if (!sceneStates.value!.measurement.lines.find((l) => l.id === id)) {
        delete lineColors[id];
      }
    }
  },
);

function handleModelError(msg: string) {
  failedMessage.value = msg;
  isFailedDialogOpen.value = true;
}

function handleFailCloseAll() {
  isFailedDialogOpen.value = false;
}

function handleGoBack() {
  const projectId = route.params.projectId;

  router.push(`/projects/${projectId}`);
}

// function logRendererMemory(tag = "") {
//   const renderer = sceneStates.value?.tresContext.value?.renderer
//     .instance as any;

//   if (!renderer) {
//     console.warn("Renderer not ready");
//     return;
//   }

//   // flush internal render list caches
//   renderer.renderLists?.dispose?.();

//   const info = renderer.info;

//   console.group(`THREE MEMORY ${tag}`);

//   console.log("Geometries:", info.memory.geometries);
//   console.log("Textures:", info.memory.textures);

//   // shader programs
//   console.log("Programs:", info.programs?.length ?? "unknown");

//   console.log("Render Calls:", info.render.calls);
//   console.log("Triangles:", info.render.triangles);
//   console.log("Lines:", info.render.lines);
//   console.log("Points:", info.render.points);

//   console.groupEnd();
// }

// onMounted(() => {
//   (window as any).memcheck = logRendererMemory;
// });

const isShowingCamDirection = computed(() => {
  return (
    selectedCamId.value != null &&
    sceneStates.value!.currentCamId.value != selectedCamId.value
  );
});
</script>

<template>
  <FailDialog
    v-model:open="isFailedDialogOpen"
    :message="failedMessage"
    icon="fa fa-times-circle"
    @close-all="handleFailCloseAll"
    @go-back="handleGoBack"
  />
  <ClientOnly>
    <div
      class="h-full relative flex flex-col justify-center items-center overflow-hidden"
    >
      <div
        class="w-full h-full absolute z-3 pointer-events-none flex justify-between"
        :class="
          sceneStates!.aspectMarginType.value == 'horizontal'
            ? 'flex-col'
            : 'row'
        "
      >
        <div
          :style="{
            width: sceneStates!.aspectMargin.width ?? '0px',
            height: sceneStates!.aspectMargin.height ?? '0px',
          }"
          class="align-start pointer-events-auto"
        ></div>
        <div
          :style="{
            width: sceneStates!.aspectMargin.width ?? '0px',
            height: sceneStates!.aspectMargin.height ?? '0px',
          }"
          class="align-end pointer-events-auto"
        ></div>
      </div>

      <div
        id="camera-props"
        class="absolute select-none max-w-50 top-0 right-0 z-10 text-white flex flex-col p-4 bg-black/20 backdrop-blur-sm rounded-bl-lg"
      >
        <p
          class="text-center w-full mb-2 font-bold border-b border-white/20 truncate"
          :class="{
            'cursor-pointer': sceneStates!.currentCamId.value != null,
          }"
          @click="selectCurrentCamShortcut"
        >
          {{
            sceneStates!.currentCamId.value == null
              ? "Spectator"
              : sceneStates!.currentCam.value.name
          }}
        </p>

        <div
          v-for="axis in ['x', 'y', 'z'] as const"
          :key="`pos-${axis}`"
          class="flex items-center gap-2 mb-1"
        >
          <p class="w-4">{{ axis }}:</p>
          <AdjustableInput
            v-model="sceneStates!.currentCam.value.position[axis]"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>

        <hr class="my-2 border-white/10" />

        <div
          v-for="axis in ['x', 'y', 'z'] as const"
          :key="`rot-${axis}`"
          class="flex items-center gap-2 mb-1"
        >
          <p class="w-4">
            θ<sub>{{ axis }}</sub
            >:
          </p>
          <AdjustableInput
            v-model="sceneStates!.currentCam.value.rotation[axis]"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
            :max="axis === 'x' ? Math.PI / 2 - 0.01 : undefined"
            :min="axis === 'x' ? -Math.PI / 2 + 0.01 : undefined"
          />
        </div>
      </div>

      <LazyMinimap :show="isMapOpen" />

      <div
        :ref="sceneStates!.tresCanvasParent"
        :style="{
          width: (sceneStates!.screenSize.width ?? 0) + 'px',
          height: (sceneStates!.screenSize.height ?? 0) + 'px',
        }"
        class="relative"
      >
        <div class="scene-wrapper">
          <MiniCameraScene v-if="sceneStates!.miniScene.visible" />
        </div>
        <CameraDirection
          v-if="selectedCam"
          :target-pos="selectedCam.position"
          label="Selected Camera"
          :show="isShowingCamDirection"
        />

        <TresCanvas
          id="canvas"
          ref="canvas"
          :window-size="false"
          clear-color="#0E0C29"
          tabindex="0"
          alpha
        >
          <LineMeasurement
            ref="lineMeasurement"
            :perspective-camera="perspectiveCamera"
            :get-surface-hit="getSurfaceHit"
          />

          <TresPerspectiveCamera
            ref="perspectiveCamera"
            :position="
              sceneStates!.transformingInfo.value?.position ??
              sceneStates!.currentCam.value?.position
            "
            :rotation="
              sceneStates!.transformingInfo.value?.rotation ??
              sceneStates!.currentCam.value?.rotation
            "
            :fov="
              sceneStates!.transformingInfo.value?.fov ??
              sceneStates!.currentCam.value?.fov
            "
            :aspect="aspect"
          />

          <CoverageAreaMesh
            v-if="isPreviewing"
            face-id="__preview__"
            color="#22ff88"
            :preview-points="previewPoints"
            :selected="true"
            :show-corners="false"
            :opacity="0.12"
            :y-offset="COVERAGE_Y_OFFSET"
            :workspace="props.workspace"
          />

          <template v-if="sceneStates!.selectionMode.value === 'coverage-area'">
            <TresMesh
              v-for="(point, i) in draftPointMarkers"
              :key="`draft-point-${i}`"
              :position="point"
              :render-order="1002"
            >
              <TresSphereGeometry :args="[0.025, 16, 16]" />
              <TresMeshBasicMaterial
                color="#ffffff"
                :transparent="true"
                :opacity="0.95"
                :depth-test="false"
                :depth-write="false"
              />
            </TresMesh>
          </template>

          <CameraObject
            v-for="[camId, cam] in Object.entries(
              sceneStates!.optimization?.candidateCameras ?? {},
            )"
            :key="camId"
            :cam-id="camId"
            :name="cam.name"
            :instance="cam"
            :workspace="props.workspace"
            color="#62B2F5"
          />

          <CameraObject
            v-for="[camId, cam] in Object.entries(sceneStates!.cameras)"
            :key="camId"
            :cam-id="camId"
            :name="cam.name"
            :workspace="props.workspace"
          />

          <FrustumOverlay />

          <!-- <AxisGizmo /> -->

          <Suspense><Environment preset="city" /></Suspense>
          <TresAmbientLight :intensity="0.4" />
          <TresDirectionalLight :position="[10, 10, 5]" :intensity="1" />

          <CalibrationGrid
            v-if="currentPanel == 'calibration' && props.workspace"
            :initial-pos="[
              sceneStates!.currentCam.value.position.x,
              sceneStates!.currentCam.value.position.y,
              sceneStates!.currentCam.value.position.z,
            ]"
            :workspace="props.workspace"
          />

          <Suspense>
            <ModelLoader
              v-if="modelPath != undefined"
              :path="modelPath!.href"
              @err="handleModelError"
            />
          </Suspense>

          <!-- Grid  1 unit = 1 m -->
          <Grid
            :position="[0, -sceneStates!.calibration.heightOffset, 0]"
            :args="[1, 1]"
            :cell-size="0.2"
            cell-color="#90EE90"
            section-color="white"
            :infinite-grid="true"
            :side="DoubleSide"
            :scale="[
              1 / sceneStates!.calibration.scale,
              1,
              1 / sceneStates!.calibration.scale,
            ]"
          />

          <template v-for="[id, face] of selectedFaces" :key="id">
            <CoverageAreaMesh
              v-model.points="sceneStates!.facesManagement.faces[id]!"
              :face-id="id"
              :color="face.color ?? '#22ff88'"
              :selected="true"
              :show-corners="false"
              :y-offset="COVERAGE_Y_OFFSET"
              :workspace="props.workspace"
            />

            <template v-if="props.workspace == 'me'">
              <CoverageCornerGizmo
                v-for="(p, index) in face.points"
                :key="`${id}-${index}`"
                :face-id="id"
                :corner-index="index"
                :position="new Vector3(p[0], p[1] + COVERAGE_Y_OFFSET, p[2])"
                :size="0.14"
                :y-offset="COVERAGE_Y_OFFSET"
              />
            </template>
          </template>
        </TresCanvas>

        <div
          id="measurement-labels"
          class="absolute inset-0 pointer-events-none"
        >
          <div
            v-for="line in sceneStates!.measurement.lines"
            :key="`label-${line.id}`"
            class="absolute z-20 pointer-events-none"
            :style="labelStyles[line.id]"
          >
            <div
              class="px-2 py-1 rounded bg-black/70 text-white text-xs whitespace-nowrap border border-white/20"
            >
              {{ line.label }}
            </div>
          </div>
        </div>
        <div :ref="sceneStates!.tresCanvasParent" class="relative">
          <AxisGizmo />
        </div>
      </div>
    </div>
  </ClientOnly>
</template>

<style scoped>
#canvas {
  height: 100%;
  width: 100%;
  min-height: 0;
}
#camera-props {
  text-shadow:
    -1px -1px 0 black,
    1px -1px 0 black,
    -1px 1px 0 black,
    1px 1px 0 black;
}

.minimap-container {
  width: 220px;
  user-select: none;
}
.minimap-container canvas {
  width: 220px;
  height: 220px;
  background: rgba(0, 0, 0, 0.5);
}

/* CUSTOM SLIDER STYLING */
.slider {
  appearance: none;
  width: 100%;
  height: 4px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
  outline: none;
}
.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 14px;
  height: 14px;
  background: #4a90e2;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid white;
}
</style>
