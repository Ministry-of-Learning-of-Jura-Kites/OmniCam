<script setup lang="ts">
import { inject, computed, ref, onMounted, onUnmounted } from "vue";

import { PerspectiveCamera, type Scene, WebGLRenderer } from "three";

import { EffectComposer } from "three/addons/postprocessing/EffectComposer.js";
import { OutputPass } from "three/addons/postprocessing/OutputPass.js";
import { RenderPass } from "three/addons/postprocessing/RenderPass.js";
import { ShaderPass } from "three/examples/jsm/postprocessing/ShaderPass.js";

import { SCENE_STATES_KEY } from "~/constants/state-keys";

const sceneStates = inject(SCENE_STATES_KEY);

if (!sceneStates) {
  throw new Error("MiniCamera must be used inside SceneStates provider");
}

const containerRef = ref<HTMLDivElement | null>(null);

const rendererContainerRef = ref<HTMLDivElement | null>(null);

let renderer: WebGLRenderer | null = null;

let composer: EffectComposer | null = null;

let shaderPass: ShaderPass | null = null;

let animationFrameId: number | null = null;

const posX = ref(24);

const posY = ref(24);

const isDragging = ref(false);

let dragOffsetX = 0;

let dragOffsetY = 0;

const previewCamera = computed(() => {
  const targetCameraId = sceneStates.value?.miniScene.targetCameraId;

  if (!targetCameraId) return null;

  return sceneStates.value?.cameras[targetCameraId];
});

const previewAspectRatio = computed(() => {
  const targetCameraId = sceneStates.value?.miniScene.targetCameraId;

  if (!targetCameraId) {
    return 16 / 9;
  }

  const cam = sceneStates!.value!.cameras[targetCameraId];

  if (!cam?.widthRes || !cam?.heightRes) {
    return 16 / 9;
  }

  return cam.widthRes / cam.heightRes;
});

const FisheyeShader = {
  uniforms: {
    tDiffuse: { value: null },
    uIsFisheye: { value: false },
    uFov: { value: 60 },
    uAspectRatio: { value: 1.0 },
  },
  vertexShader: `
    varying vec2 vUv;

    void main() {
      vUv = uv;

      gl_Position =
        projectionMatrix *
        modelViewMatrix *
        vec4(position, 1.0);
    }
  `,
  fragmentShader: `
    uniform sampler2D tDiffuse;
    varying vec2 vUv;
    uniform bool uIsFisheye;
    uniform float uFov;
    uniform float uAspectRatio;
    const float PI = 3.14159265359;
    void main() {
      vec2 uv = vUv;
      if (!uIsFisheye) {
        gl_FragColor = texture2D(
          tDiffuse,
          uv
        );

        return;
      }

      vec2 p = uv * 2.0 - 1.0;
      p.x *= uAspectRatio;
      float r = length(p);
      
      if (r > 1.0) {
        discard;
      }

      float maxTheta =
        radians(uFov * 0.5);

      float theta =
        r * maxTheta;

      float phi =
        atan(p.y, p.x);

      vec3 dir = vec3(
        sin(theta) * cos(phi),
        sin(theta) * sin(phi),
        -cos(theta)
      );

      vec2 sampleUv =
        dir.xy * 0.5 + 0.5;

      gl_FragColor = texture2D(
        tDiffuse,
        sampleUv
      );
    }
  `,
};

function createPreviewCamera() {
  const cam = previewCamera.value;

  if (!cam) return null;

  const width = rendererContainerRef.value?.clientWidth ?? 1;

  const height = rendererContainerRef.value?.clientHeight ?? 1;

  const preview = new PerspectiveCamera(cam.fov, width / height, 0.1, 1000);

  preview.position.copy(cam.position);

  preview.rotation.copy(cam.rotation);

  preview.updateProjectionMatrix();

  return preview;
}

function setupComposer(scene: Scene, camera: PerspectiveCamera) {
  if (!renderer) return;

  composer = new EffectComposer(renderer);

  const renderPass = new RenderPass(scene, camera);

  composer.addPass(renderPass);

  shaderPass = new ShaderPass(FisheyeShader, "tDiffuse");

  composer.addPass(shaderPass);

  composer.addPass(new OutputPass());
}

function renderLoop() {
  const tresContext = sceneStates!.value?.tresContext?.value;

  if (
    !renderer ||
    !tresContext ||
    !rendererContainerRef.value ||
    !previewCamera.value
  ) {
    animationFrameId = requestAnimationFrame(renderLoop);

    return;
  }

  const scene = tresContext.scene as Scene;

  const preview = createPreviewCamera();

  if (!preview) {
    animationFrameId = requestAnimationFrame(renderLoop);

    return;
  }

  const selectedCamId = sceneStates!.value?.miniScene.targetCameraId;

  const camSettings =
    selectedCamId != null ? sceneStates!.value!.cameras[selectedCamId] : null;

  const renderWidth = rendererContainerRef.value.clientWidth;
  const renderHeight = rendererContainerRef.value.clientHeight;

  const aspect =
    camSettings?.widthRes && camSettings?.heightRes
      ? camSettings.widthRes / camSettings.heightRes
      : renderWidth / renderHeight;

  preview.aspect = aspect;
  preview.updateProjectionMatrix();

  renderer.setSize(renderWidth, renderHeight, false);

  if (!composer) {
    setupComposer(scene, preview);
  }

  composer?.setSize(renderWidth, renderHeight);

  const distortion = previewCamera.value.distortion;

  const actualFov =
    selectedCamId != null
      ? sceneStates!.value!.cameras[selectedCamId]!.fov
      : preview.fov;

  if (shaderPass && distortion?.enabled) {
    shaderPass.uniforms.uFov!.value = actualFov;

    shaderPass.uniforms.uIsFisheye!.value = distortion.isFisheye;

    shaderPass.uniforms.uAspectRatio!.value = aspect;

    const renderPass = composer!.passes[0] as RenderPass;

    renderPass.camera = preview;

    composer!.render();
  } else {
    renderer.render(scene, preview);
  }

  animationFrameId = requestAnimationFrame(renderLoop);
}

function onMouseDown(e: MouseEvent) {
  if (!containerRef.value) return;

  isDragging.value = true;

  const rect = containerRef.value.getBoundingClientRect();

  dragOffsetX = e.clientX - rect.left;

  dragOffsetY = e.clientY - rect.top;
}

function onMouseMove(e: MouseEvent) {
  if (!isDragging.value) return;

  posX.value = e.clientX - dragOffsetX;

  posY.value = e.clientY - dragOffsetY;
}

function onMouseUp() {
  isDragging.value = false;
}

onMounted(() => {
  if (!rendererContainerRef.value) return;

  renderer = new WebGLRenderer({
    antialias: true,
    alpha: true,
  });

  renderer.setPixelRatio(window.devicePixelRatio);

  renderer.setClearColor(0x000000, 1);

  renderer.domElement.style.width = "100%";

  renderer.domElement.style.height = "100%";

  rendererContainerRef.value.appendChild(renderer.domElement);

  window.addEventListener("mousemove", onMouseMove);

  window.addEventListener("mouseup", onMouseUp);

  renderLoop();
});

onUnmounted(() => {
  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId);
  }

  window.removeEventListener("mousemove", onMouseMove);

  window.removeEventListener("mouseup", onMouseUp);

  composer?.dispose();

  renderer?.dispose();

  renderer?.forceContextLoss();

  renderer?.domElement.remove();
});
</script>

<template>
  <div
    ref="containerRef"
    class="mini-camera"
    :style="{
      left: `${posX}px`,
      top: `${posY}px`,
      aspectRatio: `${previewAspectRatio}`,
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

    <div class="mini-camera-body">
      <div ref="rendererContainerRef" class="renderer-container"></div>
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
  flex: 1;
  padding: 12px;
  overflow: hidden;
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
