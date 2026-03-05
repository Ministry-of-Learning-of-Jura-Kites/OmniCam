<script setup lang="ts">
import { TresCanvas, useRenderLoop } from "@tresjs/core";
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
  WebGLRenderer,
  type Scene,
  Vector3,
  Plane,
} from "three";
import { IS_MAP_OPEN_KEY, SCENE_STATES_KEY } from "@/constants/state-keys";

import { useCameraUpdate } from "./use-camera-update";
import type { IUserData } from "~/types/obj-3d-user-data";
import ModelLoader from "../model-loader/ModelLoader.vue";
import { onBeforeRouteLeave } from "vue-router";
import CoverageAreaMesh from "../coverage-area-mesh/CoverageAreaMesh.vue";
import type { CoverageFace } from "../scene-states-provider/create-scene-states"; // ปรับ path ให้ตรงโปรเจกต์
import CoverageCornerGizmos from "../coverage-area-mesh/CoverageCornerGizmos.vue";

const selectedFaces = computed<CoverageFace[]>(() => {
  return sceneStates.selectedCoverageFaces.value;
});
type SelectPlane = {
  origin: Vector3;
  normal: Vector3;
  u: Vector3;
  v: Vector3;
};

const selectPlane = ref<SelectPlane | null>(null);
const COVERAGE_Y_OFFSET = 0.01;

const _props = defineProps({
  projectId: { type: String, required: true },
  modelId: { type: String, required: true },
  workspace: { type: Object as PropType<string | null>, default: null },
});

const config = useRuntimeConfig();
const sceneStates = inject(SCENE_STATES_KEY)!;
const isMapOpen = inject(IS_MAP_OPEN_KEY)!;
const { data: modelResp } = sceneStates.modelInfo;

const modelPath = `http://${config.public.externalBackendHost}/api/v1/assets/projects/${modelResp.projectId}/models/${modelResp.modelId}/file/${modelResp.fileExtension.slice(1)}`;

// UPDATED: Starting height for the "slicing" view
const minimapHeight: Ref<number> = ref(5);

const camera = ref<PerspectiveCamera | null>(null);
const minimapCamera = ref<OrthographicCamera | null>(null);
const canvas = ref<InstanceType<typeof TresCanvas> | null>(null);
const minimapCanvas = ref<HTMLCanvasElement | null>(null);

const planeSelectionAnchor = ref<Vector3 | null>(null);
const planeSelectionPreview = ref<[number, number, number][] | null>(null);
type Point3 = [number, number, number];

const isPreviewing = computed(() => planeSelectionPreview.value?.length === 4);

const previewPoints = computed<Point3[]>(() => {
  return isPreviewing.value
    ? (planeSelectionPreview.value as Point3[])
    : ([] as Point3[]);
});

const aspect = computed(() => {
  const width = sceneStates.currentCam.value.aspectHeight;
  if (width == 0) return undefined;
  const height = sceneStates.currentCam.value.aspectHeight || 1;
  return width / height;
});

const minimapSize = 220;
const minimapFrustumSize = ref(40);

function handleMinimapZoom(event: WheelEvent) {
  const zoomSpeed = 0.05;
  const delta = event.deltaY * zoomSpeed;
  minimapFrustumSize.value = Math.max(
    5,
    Math.min(300, minimapFrustumSize.value + delta),
  );
}

// watch(minimapFrustumSize, async () => {
//   await nextTick();
//   if (minimapCamera.value) {
//     minimapCamera.value.updateProjectionMatrix();
//   }
// });

let miniCamUpdate = false;

watch(
  [minimapFrustumSize, minimapHeight, sceneStates.spectatorCameraPosition],
  () => {
    miniCamUpdate = true;
  },
  { immediate: true },
);

let miniRenderer: WebGLRenderer | null = null;

useCameraUpdate(sceneStates);

// ── Minimap Render Loop ──────────────────────────────────────────────
const { onLoop } = useRenderLoop();

onLoop(() => {
  if (
    !minimapCanvas.value ||
    !camera.value ||
    !minimapCamera.value ||
    !sceneStates.tresContext.value
  )
    return;

  const mainScene = sceneStates.tresContext.value.scene as Scene;
  const miniCam = minimapCamera.value as OrthographicCamera;

  if (!miniRenderer) {
    miniRenderer = new WebGLRenderer({
      canvas: minimapCanvas.value,
      antialias: true,
      alpha: true,
    });
    miniRenderer.setSize(minimapSize, minimapSize);
    miniRenderer.setClearColor(0x111122, 1);
  }

  if (miniCamUpdate) {
    miniCamUpdate = false;
    // CHANGE: Set the Y position to our slider value.
    // This effectively makes the camera "hover" at the slider height.
    miniCam.position.set(
      camera.value.position.x,
      minimapHeight.value,
      camera.value.position.z,
    );

    // Look straight down from that height
    miniCam.lookAt(
      camera.value.position.x,
      minimapHeight.value - 1,
      camera.value.position.z,
    );

    const halfSize = minimapFrustumSize.value / 2;
    miniCam.left = -halfSize;
    miniCam.right = halfSize;
    miniCam.top = halfSize;
    miniCam.bottom = -halfSize;

    // CHANGE: By setting near to 0.1, everything ABOVE the camera height is invisible.
    // It effectively slices the building/model at the slider's Y level.
    miniCam.near = 0.1;
    miniCam.far = 1000;

    miniCam.updateProjectionMatrix();
  }

  if (miniRenderer) {
    miniRenderer.render(mainScene, miniCam);
  }
});

onMounted(() => {
  sceneStates.spectatorPosition.refreshCameraState();
});

onUnmounted(() => {
  if (miniRenderer) miniRenderer.dispose();
});

// ── Raycasting & Input Events (Omitted same logic for brevity) ───────
const raycaster = new Raycaster();
const mouse = new Vector2();

const WORLD_UP = new Vector3(0, 1, 0);

const draggingCorner = ref<{
  faceId: string;
  cornerIndex: number;
  plane: Plane;
  pointerId: number;
} | null>(null);

function buildSnappedSelectPlane(
  hitPoint: Vector3,
  hitNormal: Vector3,
): SelectPlane {
  const cam = camera.value!;
  const camDir = new Vector3();
  cam.getWorldDirection(camDir);

  const absDotUpCam = Math.abs(camDir.dot(WORLD_UP));
  const isHorizontal = absDotUpCam > 0.75;

  if (isHorizontal) {
    const normal = WORLD_UP.clone();
    const camRight = new Vector3(1, 0, 0).applyQuaternion(cam.quaternion);
    camRight.y = 0;
    if (camRight.lengthSq() < 1e-8) camRight.set(1, 0, 0);

    const u = camRight.normalize();
    const v = new Vector3().crossVectors(normal, u).normalize();

    return { origin: hitPoint.clone(), normal, u, v };
  }

  const normal = camDir.clone();
  normal.y = 0;
  if (normal.lengthSq() < 1e-8) {
    normal.copy(hitNormal);
    normal.y = 0;
    if (normal.lengthSq() < 1e-8) normal.set(0, 0, -1);
  }
  normal.normalize();

  const v = new Vector3().crossVectors(WORLD_UP, normal).normalize();
  const u = new Vector3().crossVectors(normal, v).normalize();

  return { origin: hitPoint.clone(), normal, u, v };
}

function getPointerNDC(event: PointerEvent) {
  const ele = event.currentTarget as HTMLElement | null;
  if (!ele) return null;

  const rect = ele.getBoundingClientRect();
  if (rect.width === 0 || rect.height === 0) return null;

  return {
    x: ((event.clientX - rect.left) / rect.width) * 2 - 1,
    y: -((event.clientY - rect.top) / rect.height) * 2 + 1,
  };
}

function getSurfaceHit(
  event: PointerEvent,
): { point: Vector3; normal: Vector3 } | null {
  if (!sceneStates.tresContext.value || !camera.value) return null;

  const ndc = getPointerNDC(event);
  if (!ndc) return null;

  mouse.x = ndc.x;
  mouse.y = ndc.y;
  raycaster.setFromCamera(mouse, camera.value);

  const scene = sceneStates.tresContext.value.scene as Scene;
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

function intersectRayWithPlane(
  event: PointerEvent,
  plane: Plane,
): Vector3 | null {
  if (!sceneStates.tresContext.value || !camera.value) return null;

  const ndc = getPointerNDC(event);
  if (!ndc) return null;

  mouse.x = ndc.x;
  mouse.y = ndc.y;
  raycaster.setFromCamera(mouse, camera.value);

  const ray = raycaster.ray;
  const denom = plane.normal.dot(ray.direction);
  if (Math.abs(denom) < 1e-8) return null;

  const t = -(ray.origin.dot(plane.normal) + plane.constant) / denom;
  const hit = new Vector3();
  ray.at(t, hit);
  return hit;
}

function makeRectOnPlane(
  anchor: Vector3,
  current: Vector3,
  sp: SelectPlane,
): [number, number, number][] {
  const d = current.clone().sub(anchor);
  const du = d.dot(sp.u);
  const dv = d.dot(sp.v);

  const u0 = Math.min(0, du),
    u1 = Math.max(0, du);
  const v0 = Math.min(0, dv),
    v1 = Math.max(0, dv);

  const p00 = anchor
    .clone()
    .addScaledVector(sp.u, u0)
    .addScaledVector(sp.v, v0);
  const p10 = anchor
    .clone()
    .addScaledVector(sp.u, u1)
    .addScaledVector(sp.v, v0);
  const p11 = anchor
    .clone()
    .addScaledVector(sp.u, u1)
    .addScaledVector(sp.v, v1);
  const p01 = anchor
    .clone()
    .addScaledVector(sp.u, u0)
    .addScaledVector(sp.v, v1);
  return [
    [p00.x, p00.y, p00.z],
    [p10.x, p10.y, p10.z],
    [p11.x, p11.y, p11.z],
    [p01.x, p01.y, p01.z],
  ];
}

function computeMetricsOnPlane(
  anchor: Vector3,
  current: Vector3,
  sp: SelectPlane,
) {
  const d = current.clone().sub(anchor);
  const du = d.dot(sp.u);
  const dv = d.dot(sp.v);

  const width = Math.abs(du);
  const height = Math.abs(dv);

  // center = anchor + u*(du/2) + v*(dv/2)
  const centerV = anchor
    .clone()
    .addScaledVector(sp.u, du * 0.5)
    .addScaledVector(sp.v, dv * 0.5);

  return {
    width,
    height,
    center: [centerV.x, centerV.y, centerV.z] as [number, number, number],
    normal: [sp.normal.x, sp.normal.y, sp.normal.z] as [number, number, number],
  };
}

function handleCoverageAreaPointer(event: PointerEvent) {
  if (sceneStates.selectionMode.value !== "coverage-area") return;

  if (event.type === "pointermove") {
    if (planeSelectionAnchor.value && selectPlane.value) {
      const plane = new Plane(
        selectPlane.value.normal,
        -selectPlane.value.normal.dot(selectPlane.value.origin),
      );
      const p = intersectRayWithPlane(event, plane);
      if (!p) return;

      planeSelectionPreview.value = makeRectOnPlane(
        planeSelectionAnchor.value,
        p,
        selectPlane.value,
      );
      sceneStates.tresContext.value?.invalidate?.();
    }
    return;
  }

  if (event.type !== "pointerdown") return;

  // ---- click #1 ----
  if (!planeSelectionAnchor.value) {
    const hit = getSurfaceHit(event);
    if (!hit) return;
    const sp = buildSnappedSelectPlane(hit.point, hit.normal);
    selectPlane.value = sp;

    planeSelectionAnchor.value = sp.origin.clone();
    planeSelectionPreview.value = makeRectOnPlane(sp.origin, sp.origin, sp);
    sceneStates.tresContext.value?.invalidate?.();
    return;
  }

  // ---- click #2 ----
  if (!selectPlane.value) return;

  const plane = new Plane(
    selectPlane.value.normal,
    -selectPlane.value.normal.dot(selectPlane.value.origin),
  );
  const current = intersectRayWithPlane(event, plane);
  if (!current) return;

  const quad = makeRectOnPlane(
    planeSelectionAnchor.value,
    current,
    selectPlane.value,
  );
  const { center, width, height, normal } = computeMetricsOnPlane(
    planeSelectionAnchor.value,
    current,
    selectPlane.value,
  );

  if (width >= 0.05 && height >= 0.05) {
    sceneStates.addCoverageFace({
      id: `area_${Date.now()}`,
      points: quad,
      center,
      width,
      height,
      normal,
    });
  }

  planeSelectionAnchor.value = null;
  planeSelectionPreview.value = null;
  selectPlane.value = null;
  sceneStates.tresContext.value?.invalidate?.();
}

function onCanvasPointer(event: PointerEvent) {
  if (!sceneStates.tresContext.value) return;
  if (sceneStates.selectionMode.value === "coverage-area") {
    handleCoverageAreaPointer(event);
    return;
  }
  if (draggingCorner.value) {
    if (event.type === "pointermove") {
      const pElev = intersectRayWithPlane(event, draggingCorner.value.plane);
      if (!pElev) return;
      const p = pElev.clone();
      p.y -= COVERAGE_Y_OFFSET;

      sceneStates.updateCoverageFaceCorner(
        draggingCorner.value.faceId,
        draggingCorner.value.cornerIndex,
        [p.x, p.y, p.z],
      );
      sceneStates.tresContext.value?.invalidate?.();
      return;
    }

    if (event.type === "pointerup" || event.type === "pointercancel") {
      draggingCorner.value = null;
      return;
    }
  }

  const ndc = getPointerNDC(event);
  if (!ndc) return;
  if (!camera.value) return;
  mouse.x = ndc.x;
  mouse.y = ndc.y;
  raycaster.setFromCamera(mouse, camera.value);

  if (event.type === "pointerdown") {
    const scene = sceneStates.tresContext.value.scene as Scene;
    const hits = raycaster.intersectObjects(scene.children, true);
    const cornerHit = hits.find(
      (h) => h.object?.userData?.kind === "coverage-corner",
    );

    if (cornerHit) {
      const { faceId, cornerIndex } = cornerHit.object.userData as {
        faceId: string;
        cornerIndex: number;
      };
      const face = sceneStates.selectedCoverageFaces.value.find(
        (f) => f.id === faceId,
      );
      if (!face) return;

      const point = face.points[cornerIndex];
      if (!point) return;
      const baseP = new Vector3(...point);
      const elevatedY = baseP.y + COVERAGE_Y_OFFSET;
      const plane = new Plane(WORLD_UP.clone(), -elevatedY);

      (event.currentTarget as HTMLElement)?.setPointerCapture(event.pointerId);

      draggingCorner.value = {
        faceId,
        cornerIndex,
        plane,
        pointerId: event.pointerId,
      };

      sceneStates.tresContext.value?.invalidate?.();
      return;
    }
  }
  if (event.type === "pointerup" || event.type === "pointercancel") {
    const t = event.currentTarget as HTMLElement | null;
    if (draggingCorner.value) {
      t?.releasePointerCapture?.(draggingCorner.value.pointerId);
    }
    draggingCorner.value = null;
    return;
  }

  const ele = sceneStates.tresContext.value.renderer.domElement;
  const rect = ele.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
  raycaster.setFromCamera(mouse, camera.value!);

  const intersects = raycaster.intersectObjects(
    [...sceneStates.draggableObjects],
    false,
  );

  if (intersects.length > 0) {
    const foundObj = intersects[0];
    const userData = foundObj?.object.userData as IUserData;
    userData.handleEvent.call(userData, event.type, event);
  } else if (event.type === "pointerdown") {
    sceneStates.spectatorRotation.onPointerDown(event);
  }
}

watch(
  () => sceneStates.selectionMode.value,
  (mode) => {
    if (mode !== "coverage-area") {
      planeSelectionAnchor.value = null;
      planeSelectionPreview.value = null;
      selectPlane.value = null;
    }
  },
);

watch(
  () => sceneStates.selectedCoverageFaces.value.length,
  (len) => {
    if (len === 0) {
      draggingCorner.value = null;
      planeSelectionAnchor.value = null;
      planeSelectionPreview.value = null;
      selectPlane.value = null;
      sceneStates.tresContext.value?.invalidate?.();
    }
  },
);

watch(
  canvas,
  (newCanvas) => {
    if (!newCanvas) return;
    const context = newCanvas.context!;
    const renderer = context.renderer;
    sceneStates.tresContext.value = context;
    renderer.value.domElement.addEventListener("pointerdown", onCanvasPointer);
    renderer.value.domElement.addEventListener("pointermove", onCanvasPointer);
    renderer.value.domElement.addEventListener("pointerup", onCanvasPointer);
    renderer.value.domElement.addEventListener(
      "keydown",
      sceneStates.spectatorPosition.onKeyDown,
    );
    renderer.value.domElement.addEventListener(
      "keyup",
      sceneStates.spectatorPosition.onKeyUp,
    );
    renderer.value.domElement.addEventListener("blur", (e: FocusEvent) => {
      sceneStates.spectatorRotation.onBlur(e);
      sceneStates.spectatorPosition.onBlur(e);
    });
  },
  { once: true },
);

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

watch(
  () => [sceneStates.transformingInfo, sceneStates.currentCam],
  ([transform, cam]) => {
    const newFov = transform?.value?.fov ?? cam?.value?.fov;
    if (camera.value && newFov !== undefined) {
      camera.value.fov = newFov;
      camera.value.updateProjectionMatrix();
    }
  },
  { deep: true },
);

const handleBeforeUnload = (event: BeforeUnloadEvent) => {
  if (sceneStates.markedForCheck.size > 0) {
    const message =
      "You have unsaved camera changes. Are you sure you want to leave?";
    event.preventDefault();
    event.returnValue = message;
    return message;
  }
};
onMounted(() => {
  window.addEventListener("beforeunload", handleBeforeUnload);
});
onUnmounted(() => {
  window.removeEventListener("beforeunload", handleBeforeUnload);
});
onBeforeRouteLeave((to, from, next) => {
  if (sceneStates.markedForCheck.size > 0) {
    const answer = window.confirm(
      "You have unsaved camera changes. Are you sure you want to leave?",
    );
    if (!answer) {
      next(false);
      return;
    }
  }
  next();
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
        class="absolute top-0 right-0 z-10 text-white flex flex-col p-4 bg-black/20 backdrop-blur-sm rounded-bl-lg"
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

      <div
        v-show="isMapOpen"
        class="minimap-container absolute bottom-4 right-4 z-20 pointer-events-auto select-none"
        @wheel.prevent="handleMinimapZoom"
      >
        <div class="flex flex-col gap-1 mb-2">
          <label class="text-[10px] text-white opacity-70 uppercase">
            Cut Height: {{ minimapHeight }}m
          </label>
          <input
            id="minimap-slider"
            v-model.number="minimapHeight"
            type="range"
            min="0"
            max="5"
            step="0.01"
            class="slider"
          />
        </div>
        <div
          class="minimap-label text-white text-[10px] text-center mb-1 opacity-80 uppercase tracking-widest"
        >
          Minimap (Scroll to Zoom)
        </div>
        <canvas
          ref="minimapCanvas"
          class="rounded-lg shadow-2xl border border-gray-600/50 cursor-zoom-in bg-black"
        ></canvas>
      </div>

      <div
        :ref="sceneStates.tresCanvasParent"
        :style="{
          width: sceneStates.screenSize.width + 'px',
          height: sceneStates.screenSize.height + 'px',
        }"
        class="relative"
      >
        <TresCanvas
          id="canvas"
          ref="canvas"
          :window-size="false"
          clear-color="#0E0C29"
          tabindex="0"
          render-mode="always"
        >
          <TresPerspectiveCamera
            ref="camera"
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

          <TresOrthographicCamera ref="minimapCamera" :manual="true" />

          <CameraObject
            v-for="[camId, cam] in Object.entries(sceneStates.cameras)"
            :key="camId"
            :cam-id="camId"
            :name="cam.name"
            :workspace="workspace"
          />

          <Suspense><Environment preset="city" /></Suspense>
          <TresAmbientLight :intensity="0.4" />
          <TresDirectionalLight :position="[10, 10, 5]" :intensity="1" />

          <Suspense>
            <ModelLoader :path="modelPath" :position="[0, 0, 0]" />
          </Suspense>

          <Grid
            :args="[20, 20]"
            :cell-size="1"
            :infinite-grid="true"
            :side="DoubleSide"
          />

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

          <template v-for="face in selectedFaces" :key="face.id">
            <CoverageAreaMesh
              :face-id="face.id"
              :points="face.points"
              color="#22ff88"
              :selected="true"
              :show-corners="false"
              :y-offset="COVERAGE_Y_OFFSET"
            />

            <CoverageCornerGizmos
              :face-id="face.id"
              :points="face.points"
              :size="0.14"
              :y-offset="COVERAGE_Y_OFFSET"
              :visible="sceneStates.selectionMode.value !== 'coverage-area'"
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
