<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import AdjustableInput from "../../adjustable-input/AdjustableInput.vue";
import { SPECTATOR_ADJ_INPUT_SENTIVITY } from "~/constants";
import CameraObject from "../camera-object/CameraObject.vue";
import {
  type PerspectiveCamera,
  Raycaster,
  Vector2,
  DoubleSide,
  Vector3,
  WebGLCubeRenderTarget,
  LinearFilter,
  type CubeCamera,
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
// import Distortion from "@/components/3d/distortion/Distortion.vue";
import CubeDistortion from "@/components/3d/distortion/CubeDistortion.vue";
import CoverageAreaMesh from "../coverage-area-mesh/CoverageAreaMesh.vue";
import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";
import CoverageCornerGizmo from "../coverage-area-mesh/CoverageCornerGizmo.vue";
import { averageVector } from "~/utils/face-helper/avg-vec";
import { v4 as uuidv4 } from "uuid";
import { get3dModelPathClient } from "~/composables/api/use-fetch-model-api";
import AxisGizmo from "../axis-gizmo/axis-gizmo.vue";
import type {
  QuadrilateralPoints,
  QuadrilateralVectors,
} from "~/types/trapezoid";
import CameraDirection from "../camera-direction/CameraDirection.vue";

const { isPanelOpen, currentPanel, camPanelInfo } = inject(PANEL_KEY)!;
const { selectedCamId } = camPanelInfo;

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
const cubeCamera: Ref<CubeCamera | null> = ref(null);
// const camera = ref<PerspectiveCamera | null>(null);

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

const isPreviewing = computed(() => previewPoints.value.length > 2);

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

usePromptUnsaved(sceneStates.value!);

useCameraUpdate(sceneStates.value!);

// ── Raycasting & Input Events (Omitted same logic for brevity) ───────
const raycaster = new Raycaster();
const mouse = new Vector2();

function clearDraftCoverageSelection() {
  draftCoveragePoints.value = [];
  // sceneStates.value!.tresContext.value?.invalidate?.();
}

function buildDraftCoveragePreview(points: Vector3[]): Point3[] {
  const count = points.length;
  if (count < 1) return [];

  // 1 & 2 points: Just the raw points (renders as a point or a single line)
  if (count < 3) {
    return points.map((p) => [p.x, p.y, p.z] as Point3);
  }

  const p0 = points[0]!;
  const p1 = points[1]!;
  const p2 = points[2]!;

  // 3 points EXACTLY: Return a triangle (P0 -> P1 -> P2)
  if (count === 3) {
    return points.map((p) => [p.x, p.y, p.z] as Point3);
  }

  // 4 points: Snap the 4th point to the parallel rail
  const p3Raw = points[3]!;

  // Direction comes from the edge p1 -> p2
  const dir = new Vector3().subVectors(p2, p1).normalize();

  // Project the mouse position (p3Raw) onto the rail starting at p0
  const v = new Vector3().subVectors(p3Raw, p0);
  const dot = v.dot(dir);
  const p3 = p0.clone().add(dir.multiplyScalar(dot));

  return [p0, p1, p2, p3].map((p) => [p.x, p.y, p.z] as Point3);
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

function buildCoverageFaceFromPickedPoints(
  points: QuadrilateralVectors,
): ProcessedCoverageFace | null {
  if (points.length !== 4) return null;

  const p0 = points[0]!;
  const p1 = points[1]!;
  const p2 = points[2]!;
  const p3Raw = points[3]!;

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

    // sceneStates.value!.tresContext.value?.invalidate?.();
    return true;
  }

  // sceneStates.value!.tresContext.value?.invalidate?.();
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
  if (sceneStates.value!.selectionMode.value === "coverage-area") {
    const handled = handleCoverageAreaPointer(event);
    if (handled) return;
  }
  const ele = sceneStates.value!.tresContext.value.renderer.instance.domElement;
  const rect = ele.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height!) * 2 + 1;
  raycaster.setFromCamera(mouse, perspectiveCamera.value!);
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

const cubeCameraTarget = new WebGLCubeRenderTarget(1024, {
  generateMipmaps: true,
  minFilter: LinearFilter,
});

onMounted(() => {
  const stopPersWatch = watch(
    perspectiveCamera,
    (camera) => {
      if (camera != undefined) {
        sceneStates.value!.perspectiveCamera.value = camera;
        stopPersWatch();
      }
    },
    { immediate: true },
  );

  watch(
    cubeCamera,
    (camera) => {
      if (camera != null) {
        camera.renderTarget = cubeCameraTarget;
        sceneStates.value!.cubeCamera.value = camera;
        camera.rotation.order = "YXZ";
        watch(
          () => sceneStates.value!.currentCam.value.position.x,
          (x) => {
            camera.position.x = x;
          },
        );
        watch(
          () => sceneStates.value!.currentCam.value.position.y,
          (y) => {
            camera.position.y = y;
          },
        );
        watch(
          () => sceneStates.value!.currentCam.value.position.z,
          (z) => {
            camera.position.z = z;
          },
        );
        watch(
          () => sceneStates.value!.currentCam.value.rotation.x,
          (x) => {
            camera.rotation.x = x;
          },
        );
        watch(
          () => sceneStates.value!.currentCam.value.rotation.y,
          (y) => {
            camera.rotation.y = y;
          },
        );
        watch(
          () => sceneStates.value!.currentCam.value.rotation.z,
          (z) => {
            camera.rotation.z = z;
          },
        );
      }
    },
    { immediate: true },
  );

  watch(
    () => canvas.value?.context,
    (context) => {
      if (!context) return;
      const renderer = context.renderer;
      let frameDurations: number[] = [];
      const WINDOW_SIZE = 60;

      // O(1) Global Tracking for P95
      let p95GlobalAverage = 0;
      let p95SampleCount = 0;

      // O(1) Global Tracking for every single Frame
      let globalFrameAvg = 0;
      let totalFrameCount = 0;

      setInterval(() => {
        if (frameDurations.length >= WINDOW_SIZE) {
          const sorted = [...frameDurations].sort((a, b) => a - b);
          const p95 = sorted[Math.floor(0.95 * (WINDOW_SIZE - 1))]!;

          const windowAvg =
            frameDurations.reduce((a, b) => a + b, 0) / WINDOW_SIZE;

          // 2. Update Global P95 Average (O(1) space)
          p95SampleCount++;
          p95GlobalAverage =
            (p95 + p95GlobalAverage * (p95SampleCount - 1)) / p95SampleCount;

          // Local copies for the async log
          const snapshot = {
            windowAvg,
            p95,
            p95Global: p95GlobalAverage,
            frameGlobal: globalFrameAvg,
            totalFrames: totalFrameCount,
          };

          console.log(
            `[Stats] WinAvg: ${snapshot.windowAvg.toFixed(2)}ms | ` +
              `%cP95: ${snapshot.p95.toFixed(2)}ms%c | ` +
              `Global P95: %c${snapshot.p95Global.toFixed(2)}ms%c | ` +
              `Global Frame: %c${snapshot.frameGlobal.toFixed(3)}ms`,
            "color: #ffaa00; font-weight: bold;", // P95
            "color: inherit;",
            "color: #00d4ff; font-weight: bold;", // Global P95
            "color: inherit;",
            "color: #00ff00; font-weight: bold;", // Global Frame
          );

          frameDurations = [];
        }
      });

      renderer.loop.onBeforeLoop((time: { delta: number; elapsed: number }) => {
        const delta = time.delta * 1000; // Assuming time.delta is in seconds

        // 1. Update Global Frame Average (O(1) space)
        // This updates every single frame for maximum precision
        totalFrameCount++;
        globalFrameAvg =
          (delta + globalFrameAvg * (totalFrameCount - 1)) / totalFrameCount;

        frameDurations.push(delta);
        // stats!.begin();
      });
      // let frameCount = 0;
      renderer.loop.onLoop(() => {
        // stats!.end();
        // if (frameCount % 100 === 0) {
        //   console.table({
        //     // eslint-disable-next-line @typescript-eslint/no-explicit-any
        //     "Draw Calls": (renderer.instance as any).info.render.calls,
        //     Triangles: renderer.instance.info.render.triangles,
        //     "Geometries (Mem)": renderer.instance.info.memory.geometries,
        //     "Textures (Mem)": renderer.instance.info.memory.textures,
        //   });
        // }
        // frameCount++;
      });
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
    // stats.showPanel(1);
    // stats.showPanel(2); // 0: fps, 1: ms, 2: mb, 3+: custom
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

const isShowingCamDirection = computed(() => {
  return (
    selectedCamId.value != null &&
    sceneStates.value!.currentCamId.value != selectedCamId.value
  );
});
</script>

<template>
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
          :preserve-drawing-buffer="true"
          clear-color="#0E0C29"
          tabindex="0"
          alpha
        >
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

          <TresCubeCamera ref="cubeCamera" />

          <!-- <Distortion /> -->
          <CubeDistortion />

          <!-- <TresMesh>
            <TresBoxGeometry :args="[2, 2, 2, 32, 32, 32]" />
            <TresMeshStandardMaterial
              :wireframe="true"
              @before-compile="injectFisheye"
            />
          </TresMesh> -->
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

          <AxisGizmo />

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
