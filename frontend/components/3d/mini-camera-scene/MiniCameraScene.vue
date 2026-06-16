<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from "vue";
import { PerspectiveCamera, type Scene, WebGLRenderer } from "three";
import { SCENE_STATES_KEY } from "~/constants/state-keys";
import { createCubeDistortionRenderer } from "~/composables/use-cube-distortion";

const sceneStates = inject(SCENE_STATES_KEY);

if (!sceneStates) {
  throw new Error("MiniCamera must be used inside SceneStates provider");
}

const containerRef = ref<HTMLDivElement | null>(null);

const bodyRef = ref<HTMLDivElement | null>(null);

const rendererContainerRef = ref<HTMLDivElement | null>(null);

let renderer: WebGLRenderer | null = null;
let animationFrameId: number | null = null;
let distortion: ReturnType<typeof createCubeDistortionRenderer> | null = null;

const standardCamera = new PerspectiveCamera();
const posX = ref(24);
const posY = ref(24);

const isDragging = ref(false);

let dragOffsetX = 0;
let dragOffsetY = 0;

const BODY_WIDTH = 520;

const previewCamera = computed(() => {
  const targetCameraId = sceneStates.value?.miniScene.targetCameraId;

  return targetCameraId ? sceneStates.value?.cameras[targetCameraId] : null;
});

const previewAspectRatio = computed(() => {
  const targetCameraId = sceneStates.value?.miniScene.targetCameraId;

  if (!targetCameraId) {
    return 16 / 9;
  }

  const cam = sceneStates.value!.cameras[targetCameraId];

  return cam?.widthRes && cam?.heightRes
    ? cam.widthRes / cam.heightRes
    : 16 / 9;
});

const bodyHeight = computed(() => {
  return BODY_WIDTH / previewAspectRatio.value;
});

function onMouseDown(e: MouseEvent) {
  if (!containerRef.value) {
    return;
  }

  isDragging.value = true;

  dragOffsetX = e.clientX - posX.value;
  dragOffsetY = e.clientY - posY.value;
}

function onMouseMove(e: MouseEvent) {
  if (!isDragging.value) {
    return;
  }

  posX.value = e.clientX - dragOffsetX;
  posY.value = e.clientY - dragOffsetY;
}

function onMouseUp() {
  isDragging.value = false;
}

watch(
  () => [
    previewCamera.value?.distortion?.enabled,
    previewCamera.value?.distortion?.isFisheye,
  ],

  () => {
    distortion?.dispose();
    distortion = null;
  },
);

function updateStandardCamera(aspectRatio: number) {
  if (!previewCamera.value) {
    return;
  }

  standardCamera.fov = previewCamera.value.fov;
  standardCamera.aspect = aspectRatio;
  standardCamera.near = 0.1;
  standardCamera.far = 1000;

  standardCamera.position.copy(previewCamera.value.position);
  standardCamera.rotation.copy(previewCamera.value.rotation);

  standardCamera.updateProjectionMatrix();
  standardCamera.updateMatrixWorld();
}

function renderStandardView(
  scene: Scene,
  aspectRatio: number,
  renderWidth: number,
  renderHeight: number,
) {
  if (!renderer) {
    return;
  }

  updateStandardCamera(aspectRatio);
  renderer.setSize(renderWidth, renderHeight, false);
  renderer.render(scene, standardCamera);
}

function renderDistortionView(
  scene: Scene,
  aspectRatio: number,
  renderWidth: number,
  renderHeight: number,
  isFisheye: boolean,
) {
  if (!renderer) {
    return;
  }

  updateStandardCamera(aspectRatio);

  if (!distortion) {
    distortion = createCubeDistortionRenderer(renderer, scene);
  }

  distortion.render({
    position: standardCamera.position,
    rotation: standardCamera.rotation,
    fov: standardCamera.fov,
    aspectRatio,
    width: renderWidth,
    height: renderHeight,
    isFisheye,
    activeCamera: standardCamera,
  });

  renderer.resetState();
}

function renderLoop() {
  const tresContext = sceneStates!.value?.tresContext?.value;
  console.log("previewCamera", previewCamera.value);
  console.log("distortion enabled", previewCamera.value?.distortion?.enabled);
  console.log("fisheye", previewCamera.value?.distortion?.isFisheye);
  if (
    !renderer ||
    !rendererContainerRef.value ||
    !tresContext ||
    !previewCamera.value
  ) {
    animationFrameId = requestAnimationFrame(renderLoop);

    return;
  }

  const scene = tresContext.scene as Scene;
  const renderWidth = rendererContainerRef.value.clientWidth;
  const renderHeight = rendererContainerRef.value.clientHeight;
  const aspectRatio = renderWidth / renderHeight;

  const isDistortionEnabled = previewCamera.value.distortion?.enabled ?? false;
  const isFisheye = previewCamera.value.distortion?.isFisheye ?? false;

  if (isDistortionEnabled) {
    renderDistortionView(
      scene,
      aspectRatio,
      renderWidth,
      renderHeight,
      isFisheye,
    );
  } else {
    renderStandardView(scene, aspectRatio, renderWidth, renderHeight);
  }

  animationFrameId = requestAnimationFrame(renderLoop);
}

onMounted(() => {
  renderer = new WebGLRenderer({
    antialias: true,
    alpha: true,
  });
  renderer.domElement.style.width = "100%";
  renderer.domElement.style.height = "100%";
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  rendererContainerRef.value?.appendChild(renderer.domElement);
  window.addEventListener("mousemove", onMouseMove);
  window.addEventListener("mouseup", onMouseUp);

  renderLoop();
});

onUnmounted(() => {
  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId);

    animationFrameId = null;
  }

  window.removeEventListener("mousemove", onMouseMove);
  window.removeEventListener("mouseup", onMouseUp);
  distortion?.dispose();
  renderer?.dispose();
});
</script>

<template>
  <div
    ref="containerRef"
    class="mini-camera"
    :style="{
      left: `${posX}px`,
      top: `${posY}px`,
    }"
    @mousedown="onMouseDown"
  >
    <div class="mini-camera-header">
      <div class="header-left">
        <div class="live-dot"></div>

        <span class="title"> Camera Preview </span>
      </div>

      <div class="header-right">
        {{ previewCamera?.name ?? "No Camera" }}
      </div>
    </div>

    <div
      ref="bodyRef"
      class="mini-camera-body"
      :style="{
        height: `${bodyHeight}px`,
      }"
    >
      <div ref="rendererContainerRef" class="renderer-container" />
    </div>
  </div>
</template>

<style scoped>
.mini-camera {
  position: absolute;
  width: 520px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 18px;
  background: rgba(15, 15, 18, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.04),
    0 20px 50px rgba(0, 0, 0, 0.45),
    0 6px 18px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(16px);
  z-index: 1000;
  user-select: none;
  cursor: grab;
}

.mini-camera:active {
  cursor: grabbing;
}

.mini-camera-header {
  height: 52px;
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: linear-gradient(
    to bottom,
    rgba(255, 255, 255, 0.08),
    rgba(255, 255, 255, 0.02)
  );
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.live-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: #22c55e;
  box-shadow:
    0 0 10px #22c55e,
    0 0 20px #22c55e;
}

.title {
  color: white;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.header-right {
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
  font-weight: 500;
}

.mini-camera-body {
  height: 100%;
  width: 100%;
  padding: 12px;
  overflow: hidden;
  flex-shrink: 0;
}

.renderer-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: 12px;
  background: black;
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.03);
}
</style>
