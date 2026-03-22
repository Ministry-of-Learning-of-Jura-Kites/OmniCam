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
  type OrthographicCamera,
  WebGLCubeRenderTarget,
  LinearFilter,
  type CubeCamera,
  Vector3,
} from "three";
import { IS_MAP_OPEN_KEY, SCENE_STATES_KEY } from "@/constants/state-keys";
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
import type { CoverageFace } from "../scene-states-provider/create-scene-states";
import CoverageCornerGizmos from "../coverage-area-mesh/CoverageCornerGizmos.vue";

const selectedFaces = computed(() =>
  sceneStates.facesManagement.faces.value.filter(
    (face) => !sceneStates.facesManagement.isAllHidden.value && !face.hidden,
  ),
);
type Point3 = [number, number, number];
const WORLD_UP = new Vector3(0, 1, 0);
const props = defineProps({
  projectId: { type: String, required: true },
  modelId: { type: String, required: true },
  workspace: { type: String, default: null },
});

const config = useRuntimeConfig();
const sceneStates = inject(SCENE_STATES_KEY)!;
const { data: modelResp } = sceneStates.modelInfo;

const modelPath = `http://${config.public.externalBackendHost}/api/v1/assets/projects/${modelResp.projectId}/models/${modelResp.modelId}/file/${modelResp.fileExtension.slice(1)}`;

const perspectiveCamera = ref<PerspectiveCamera | null>(null);
const canvas: Ref<InstanceType<typeof TresCanvas> | null> = ref(null);
const cubeCamera: Ref<CubeCamera | null> = ref(null);
// const camera = ref<PerspectiveCamera | null>(null);

const minimapCamera = ref<OrthographicCamera | null>(null);
const isMapOpen = inject(IS_MAP_OPEN_KEY)!;

const COVERAGE_Y_OFFSET = 0.01;

const draftCoveragePoints = ref<Vector3[]>([]);
const draftCoverageNormals = ref<Vector3[]>([]);

const aspect = computed(() => {
  const width = sceneStates.currentCam.value.widthRes;
  if (width == 0) return undefined;
  const height = sceneStates.currentCam.value.heightRes || 1;
  return width / height;
});

const previewPoints = computed<Point3[]>(() => {
  if (sceneStates.facesManagement.mode.value !== "coverage-area") return [];

  return buildDraftCoveragePreview(
    draftCoveragePoints.value,
    draftCoverageNormals.value,
  );
});

const isPreviewing = computed(() => previewPoints.value.length === 4);

const draftPointMarkers = computed<Point3[]>(() => {
  if (sceneStates.facesManagement.mode.value !== "coverage-area") return [];

  return draftCoveragePoints.value.map((p) => [p.x, p.y, p.z] as Point3);
});

usePromptUnsaved(sceneStates);

useCameraUpdate(sceneStates);

// ── Raycasting & Input Events (Omitted same logic for brevity) ───────
const raycaster = new Raycaster();
const mouse = new Vector2();

function clearDraftCoverageSelection() {
  draftCoveragePoints.value = [];
  draftCoverageNormals.value = [];
  // sceneStates.tresContext.value?.invalidate?.();
}

function makeBasisFromNormal(normal: Vector3) {
  const n = normal.clone().normalize();

  let u = new Vector3().crossVectors(WORLD_UP, n);
  if (u.lengthSq() < 1e-8) {
    u = new Vector3(1, 0, 0).cross(n);
  }
  if (u.lengthSq() < 1e-8) {
    u = new Vector3(0, 0, 1);
  }
  u.normalize();

  const v = new Vector3().crossVectors(n, u).normalize();
  return { u, v };
}

function averageVector(vs: Vector3[]) {
  const out = new Vector3();
  if (vs.length === 0) return out;
  vs.forEach((v) => out.add(v));
  out.multiplyScalar(1 / vs.length);
  return out;
}

function computeStableNormal(points: Vector3[], normals: Vector3[]) {
  const avgNormal = averageVector(normals);
  if (avgNormal.lengthSq() > 1e-8) return avgNormal.normalize();

  if (points.length >= 3) {
    const a = points[0]!;
    const b = points[1]!;
    const c = points[2]!;
    const n = new Vector3()
      .subVectors(b, a)
      .cross(new Vector3().subVectors(c, a));

    if (n.lengthSq() > 1e-8) return n.normalize();
  }

  return WORLD_UP.clone();
}

function orderPointsOnPlane(points: Vector3[], normal: Vector3) {
  const center = averageVector(points);
  const { u, v } = makeBasisFromNormal(normal);

  const projected = points.map((p) => {
    const d = p.clone().sub(center);
    const du = d.dot(u);
    const dv = d.dot(v);
    return center.clone().addScaledVector(u, du).addScaledVector(v, dv);
  });

  return projected
    .map((p) => {
      const d = p.clone().sub(center);
      return {
        p,
        angle: Math.atan2(d.dot(v), d.dot(u)),
      };
    })
    .sort((a, b) => a.angle - b.angle)
    .map((x) => x.p);
}

function buildDraftCoveragePreview(
  points: Vector3[],
  normals: Vector3[],
): Point3[] {
  if (points.length < 3) return [];

  const normal = computeStableNormal(points, normals);
  const ordered = orderPointsOnPlane(points, normal);

  const padded =
    ordered.length === 3
      ? [...ordered, ordered[2]!.clone()]
      : ordered.slice(0, 4);

  return padded.map((p) => [p.x, p.y, p.z] as Point3);
}

function buildCoverageFaceFromPickedPoints(
  points: Vector3[],
  normals: Vector3[],
): CoverageFace | null {
  if (points.length !== 4) return null;

  const normal = computeStableNormal(points, normals);
  const ordered = orderPointsOnPlane(points, normal);

  const p0 = ordered[0]!;
  const p1 = ordered[1]!;
  const p2 = ordered[2]!;
  const p3 = ordered[3]!;

  const centerV = averageVector(ordered);

  const width = (p0.distanceTo(p1) + p2.distanceTo(p3)) / 2;

  const height = (p1.distanceTo(p2) + p3.distanceTo(p0)) / 2;

  if (width < 0.05 || height < 0.05) return null;

  return {
    id: `area_${Date.now()}`,
    points: [
      [p0.x, p0.y, p0.z],
      [p1.x, p1.y, p1.z],
      [p2.x, p2.y, p2.z],
      [p3.x, p3.y, p3.z],
    ],
    center: [centerV.x, centerV.y, centerV.z],
    normal: [normal.x, normal.y, normal.z],
  };
}

function handleCoverageAreaPointer(event: PointerEvent) {
  if (sceneStates.facesManagement.mode.value !== "coverage-area") return false;

  if (event.type !== "pointerdown" || !event.ctrlKey) {
    return false;
  }

  if (draftCoveragePoints.value.length >= 4) {
    clearDraftCoverageSelection();
  }

  const hit = getSurfaceHit(event);

  if (!hit) return true;

  draftCoveragePoints.value = [...draftCoveragePoints.value, hit.point.clone()];
  draftCoverageNormals.value = [
    ...draftCoverageNormals.value,
    hit.normal.clone(),
  ];

  if (draftCoveragePoints.value.length === 4) {
    const face = buildCoverageFaceFromPickedPoints(
      draftCoveragePoints.value,
      draftCoverageNormals.value,
    );

    if (face) {
      console.log("Adding coverage face", face);
      sceneStates.facesManagement.add(face);
    }

    requestAnimationFrame(() => {
      clearDraftCoverageSelection();
    });

    // sceneStates.tresContext.value?.invalidate?.();
    return true;
  }

  // sceneStates.tresContext.value?.invalidate?.();
  return true;
}

function getSurfaceHit(
  event: PointerEvent,
): { point: Vector3; normal: Vector3 } | null {
  if (!sceneStates.tresContext.value || !perspectiveCamera.value) return null;

  const ele = sceneStates.tresContext.value.renderer.instance.domElement;
  const rect = ele.getBoundingClientRect();

  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

  raycaster.setFromCamera(mouse, perspectiveCamera.value);

  const scene = sceneStates.tresContext.value.scene;
  const hits = raycaster.intersectObjects(scene.children, true);

  const hit = hits.find((h) => {
    const object = h.object as unknown as {
      isMesh?: boolean;
      userData?: Record<string, unknown>;
    };
    const isMesh = object.isMesh && !!h.face;
    const kind = String(object.userData?.kind ?? "");
    return isMesh && !kind.startsWith("coverage-");
  });

  if (!hit || !hit.face) return null;

  const worldNormal = hit.face.normal
    .clone()
    .transformDirection(hit.object.matrixWorld)
    .normalize();

  return { point: hit.point.clone(), normal: worldNormal };
}

function onCanvasPointer(event: PointerEvent) {
  if (!sceneStates.tresContext.value || !perspectiveCamera.value) return;
  if (sceneStates.facesManagement.mode.value === "coverage-area") {
    const handled = handleCoverageAreaPointer(event);
    if (handled) return;
  }
  const ele = sceneStates.tresContext.value.renderer.instance.domElement;
  const rect = ele.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height!) * 2 + 1;
  raycaster.setFromCamera(mouse, perspectiveCamera.value!);
  const intersects = raycaster.intersectObjects(
    [...sceneStates.draggableObjects],
    false,
  );
  if (intersects.length > 0) {
    const foundObj = intersects[0];
    const userData = foundObj?.object.userData as IUserData;
    userData.handleEvent.call(userData, event.type, event);
  } else if (
    event.type === "pointerdown" &&
    (sceneStates.currentCamId.value == null || props.workspace != null)
  ) {
    sceneStates.spectatorRotation.onPointerDown(event);
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
        sceneStates.perspectiveCamera.value = camera;
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
        sceneStates.cubeCamera.value = camera;
        camera.rotation.order = "YXZ";
        watch(
          () => sceneStates.currentCam.value.position.x,
          (x) => {
            camera.position.x = x;
          },
        );
        watch(
          () => sceneStates.currentCam.value.position.y,
          (y) => {
            camera.position.y = y;
          },
        );
        watch(
          () => sceneStates.currentCam.value.position.z,
          (z) => {
            camera.position.z = z;
          },
        );
        watch(
          () => sceneStates.currentCam.value.rotation.x,
          (x) => {
            camera.rotation.x = x;
          },
        );
        watch(
          () => sceneStates.currentCam.value.rotation.y,
          (y) => {
            camera.rotation.y = y;
          },
        );
        watch(
          () => sceneStates.currentCam.value.rotation.z,
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
      if (!context || !stats) return;
      const renderer = context.renderer;
      renderer.loop.onBeforeLoop(() => {
        stats!.begin();
      });
      renderer.loop.onLoop(() => {
        stats!.end();
      });
      sceneStates.tresContext.value = context;
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
      renderer.instance.domElement.addEventListener(
        "keydown",
        sceneStates.spectatorPosition.onKeyDown,
      );
      renderer.instance.domElement.addEventListener(
        "keyup",
        sceneStates.spectatorPosition.onKeyUp,
      );
      renderer.instance.domElement.addEventListener("blur", (e: FocusEvent) => {
        sceneStates.spectatorRotation.onBlur(e);
        sceneStates.spectatorPosition.onBlur(e);
      });

      renderer.instance.domElement.addEventListener(
        "contextmenu",
        (event: Event) => {
          event.preventDefault();

          sceneStates.spectatorRotation.onBlur(event as unknown as FocusEvent);
          sceneStates.spectatorPosition.onBlur(event as unknown as FocusEvent);
        },
      );
    },
    { once: true },
  );
});

const spectatorRefs = {
  position: {
    x: toRef(sceneStates.spectatorCameraPosition, "x"),
    y: toRef(sceneStates.spectatorCameraPosition, "y"),
    z: toRef(sceneStates.spectatorCameraPosition, "z"),
  },
  rotation: {
    x: toRef(sceneStates.spectatorCameraRotation, "x"),
    y: toRef(sceneStates.spectatorCameraRotation, "y"),
    z: toRef(sceneStates.spectatorCameraRotation, "z"),
  },
};

onMounted(() => {
  stats = new Stats();
  stats.showPanel(0);
  // stats.showPanel(1);
  // stats.showPanel(2); // 0: fps, 1: ms, 2: mb, 3+: custom
  if (document) {
    document.body.appendChild(stats.dom);
  }
});
watch(
  () => sceneStates.facesManagement.mode.value,
  (mode) => {
    if (mode !== "coverage-area") {
      clearDraftCoverageSelection();
    }
  },
);

watch(
  () => sceneStates.facesManagement.faces.value.length,
  (len) => {
    if (len === 0) {
      clearDraftCoverageSelection();
    }
  },
);
</script>

<template>
  <ClientOnly>
    <div
      class="h-full relative flex flex-col justify-center items-center overflow-hidden"
    >
      <div
        class="w-full h-full absolute z-3 pointer-events-none flex justify-between"
        :class="
          sceneStates.aspectMarginType.value == 'horizontal'
            ? 'flex-col'
            : 'row'
        "
      >
        <div
          :style="{
            width: sceneStates.aspectMargin.width ?? '0px',
            height: sceneStates.aspectMargin.height ?? '0px',
          }"
          class="align-start pointer-events-auto"
        ></div>
        <div
          :style="{
            width: sceneStates.aspectMargin.width ?? '0px',
            height: sceneStates.aspectMargin.height ?? '0px',
          }"
          class="align-end pointer-events-auto"
        ></div>
      </div>

      <div
        id="camera-props"
        class="absolute select-none top-0 right-0 z-10 text-white flex flex-col p-4 bg-black/20 backdrop-blur-sm rounded-bl-lg"
      >
        <p class="text-center w-full mb-2 font-bold border-b border-white/20">
          Spectator
        </p>

        <div
          v-for="axis in ['x', 'y', 'z'] as const"
          :key="`pos-${axis}`"
          class="flex items-center gap-2 mb-1"
        >
          <p class="w-4">{{ axis }}:</p>
          <AdjustableInput
            v-model="spectatorRefs.position[axis].value"
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
            v-model="spectatorRefs.rotation[axis].value"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
            :max="axis === 'x' ? Math.PI / 2 - 0.01 : undefined"
            :min="axis === 'x' ? -Math.PI / 2 + 0.01 : undefined"
          />
        </div>
      </div>

      <LazyMinimap :show="isMapOpen" :minimap-camera="minimapCamera" />

      <div
        :ref="sceneStates.tresCanvasParent"
        :style="{
          width: (sceneStates.screenSize.width ?? 0) + 'px',
          height: (sceneStates.screenSize.height ?? 0) + 'px',
        }"
        class="relative"
      >
        <TresCanvas
          id="canvas"
          ref="canvas"
          :window-size="false"
          clear-color="#0E0C29"
          tabindex="0"
        >
          <TresPerspectiveCamera
            ref="perspectiveCamera"
            :position="
              sceneStates.transformingInfo.value?.position ??
              sceneStates.currentCam.value?.position
            "
            :rotation="
              sceneStates.transformingInfo.value?.rotation ??
              sceneStates.currentCam.value?.rotation
            "
            :fov="
              sceneStates.transformingInfo.value?.fov ??
              sceneStates.currentCam.value?.fov
            "
            :aspect="aspect"
          />

          <TresCubeCamera ref="cubeCamera" />

          <!-- <Distortion /> -->
          <CubeDistortion />

          <TresOrthographicCamera ref="minimapCamera" :manual="true" />

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
            :points="previewPoints"
            color="#22ff88"
            :selected="true"
            :show-corners="false"
            :opacity="0.12"
            :y-offset="COVERAGE_Y_OFFSET"
          />

          <template
            v-if="sceneStates.facesManagement.mode.value === 'coverage-area'"
          >
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
            v-for="[camId, cam] in Object.entries(sceneStates.cameras)"
            :key="camId"
            :cam-id="camId"
            :name="cam.name"
            :workspace="props.workspace"
          />

          <FrustumOverlay />

          <Suspense><Environment preset="city" /></Suspense>
          <TresAmbientLight :intensity="0.4" />
          <TresDirectionalLight :position="[10, 10, 5]" :intensity="1" />

          <CalibrationGrid :workspace="props.workspace" />

          <Suspense>
            <ModelLoader :path="modelPath" />
          </Suspense>

          <!-- Grid  1 unit = 1 virtual m -->
          <Grid
            :position="[0, -sceneStates.calibration.heightOffset, 0]"
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

          <template v-for="face in selectedFaces" :key="face.id">
            <CoverageAreaMesh
              :face-id="face.id"
              :points="face.points"
              :color="face.color ?? '#22ff88'"
              :selected="true"
              :show-corners="false"
              :y-offset="COVERAGE_Y_OFFSET"
            />

            <CoverageCornerGizmos
              :face-id="face.id"
              :points="face.points"
              :size="0.14"
              :y-offset="COVERAGE_Y_OFFSET"
              :visible="
                sceneStates.facesManagement.mode.value !== 'coverage-area'
              "
            />
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
