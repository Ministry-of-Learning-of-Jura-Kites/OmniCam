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
} from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

import { useCameraUpdate } from "./use-camera-update";
import type { IUserData } from "~/types/obj-3d-user-data";
import ModelLoader from "../model-loader/ModelLoader.vue";
import { onBeforeRouteLeave } from "vue-router";

const props = defineProps({
  projectId: { type: String, required: true },
  modelId: { type: String, required: true },
  workspace: { type: String, default: null },
});

const config = useRuntimeConfig();
const sceneStates = inject(SCENE_STATES_KEY)!;
const { data: modelResp } = sceneStates.modelInfo;

const modelPath = `http://${config.public.externalBackendHost}/api/v1/assets/projects/${modelResp.projectId}/models/${modelResp.modelId}/file/${modelResp.fileExtension.slice(1)}`;

// UPDATED: Starting height for the "slicing" view
const minimapHeight: Ref<number> = ref(100);

const camera = ref<PerspectiveCamera | null>(null);
const minimapCamera = ref<OrthographicCamera | null>(null);
const canvas = ref<InstanceType<typeof TresCanvas> | null>(null);
const minimapCanvas = ref<HTMLCanvasElement | null>(null);

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

watch(minimapFrustumSize, async () => {
  await nextTick();
  if (minimapCamera.value) {
    minimapCamera.value.updateProjectionMatrix();
  }
});

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

  // CHANGE: Set the Y position to our slider value.
  // This effectively makes the camera "hover" at the slider height.
  miniCam.position.set(
    camera.value.position.x,
    minimapHeight.value,
    camera.value.position.z,
  );

  // Look straight down from that height
  miniCam.lookAt(camera.value.position.x, 0, camera.value.position.z);

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

function onCanvasPointer(event: PointerEvent) {
  if (!sceneStates.tresContext.value) return;
  const ele = sceneStates.tresContext.value.renderer.domElement;
  const rect = ele.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height!) * 2 + 1;
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

onMounted(() => {
  const handleBeforeUnload = (event: BeforeUnloadEvent) => {
    if (sceneStates.markedForCheck.size > 0) {
      const message =
        "You have unsaved camera changes. Are you sure you want to leave?";
      event.preventDefault();
      event.returnValue = message;
      return message;
    }
  };
  window.addEventListener("beforeunload", handleBeforeUnload);
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
    <div class="h-full relative flex flex-col justify-center items-center">
      <div
        class="w-full h-full absolute z-3 pointer-events-none flex justify-between"
        :class="
          'flex-' +
          (sceneStates.aspectMarginType.value == 'horizontal' ? 'col' : 'row')
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
        class="absolute top-0 right-0 z-10 text-white flex flex-col p-2"
      >
        <p class="text-center w-full h-full">Spectator</p>
        <div v-for="axis in ['x', 'y', 'z']" :key="axis" class="flex">
          <p>{{ axis }}:</p>
          <AdjustableInput
            v-model="spectatorRefs.position[axis as 'x' | 'y' | 'z'].value"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
      </div>

      <div
        class="minimap-container absolute bottom-4 right-4 z-20 pointer-events-auto select-none"
        @wheel.prevent="handleMinimapZoom"
      >
        <div class="flex flex-col gap-1 mb-2">
          <label class="text-[10px] text-white opacity-70 uppercase"
            >Cut Height: {{ minimapHeight }}m</label
          >
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
          class="rounded-lg shadow-2xl border border-gray-600/50 cursor-zoom-in"
        ></canvas>
      </div>

      <div
        :ref="sceneStates.tresCanvasParent"
        :style="{
          width: sceneStates.screenSize.width + 'px',
          height: sceneStates.screenSize.height + 'px',
        }"
      >
        <TresCanvas
          id="canvas"
          ref="canvas"
          :window-size="false"
          clear-color="#0E0C29"
          tabindex="0"
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

          <TresOrthographicCamera ref="minimapCamera" />

          <CameraObject
            v-for="[camId, cam] in Object.entries(sceneStates.cameras)"
            :key="camId"
            :cam-id="camId"
            :name="cam.name"
            :workspace="props.workspace"
          />

          <Suspense><Environment preset="city" /></Suspense>
          <TresAmbientLight :intensity="0.4" />
          <TresDirectionalLight :position="[10, 10, 5]" :intensity="1" />
          <Suspense
            ><ModelLoader :path="modelPath" :position="[0, 0, 0]"
          /></Suspense>

          <Grid
            :args="[20, 20]"
            :cell-size="1"
            :infinite-grid="true"
            :side="DoubleSide"
          />
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
