<script setup lang="ts">
import {
  TresCanvas,
  type TresCanvasInstance,
  type TresContext,
} from "@tresjs/core";
import type { Euler, Vector3 } from "three";
import { SpriteMaterial, Sprite, CanvasTexture, Scene } from "three";
import { MINIMAP_LAYER } from "~/constants";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const props = defineProps({
  show: { type: Boolean, default: true },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

// --- Minimap Config ---
const minimapSize = 220;
const minimapHeight = ref(5);
const minimapFrustumSize = ref(20 * sceneStates.value!.calibration.scale);

// --- Texture & Sprite Setup ---
const cursorTexture = createCursorTexture();
const spriteMaterial = new SpriteMaterial({
  map: cursorTexture,
  depthTest: false,
  depthWrite: false,
});

const cursorSprite = new Sprite(spriteMaterial);
cursorSprite.layers.set(MINIMAP_LAYER);

const overlayScene = new Scene();
overlayScene.add(cursorSprite);

const canvas = ref<TresCanvasInstance>();

// --- Zoom Logic ---
function handleMinimapZoom(event: WheelEvent) {
  const zoomSpeed = 0.05;
  minimapFrustumSize.value = Math.max(
    5,
    Math.min(300, minimapFrustumSize.value + event.deltaY * zoomSpeed),
  );
}

const CURSOR_HEIGHT_OFFSET = 0.5;

function updateCursorSize(minimapFrustumSize: number) {
  const desiredPixelSize = 40;
  const unitsPerPixel = minimapFrustumSize / minimapSize;
  const unitScale = desiredPixelSize * unitsPerPixel;
  cursorSprite.scale.set(unitScale, unitScale, 1);
}

function updateCursorDepth(minimapHeight: number) {
  cursorSprite.position.y = minimapHeight - CURSOR_HEIGHT_OFFSET;
}

function updateCursor(pos: Vector3, rot: Euler) {
  cursorSprite.position.set(
    pos.x,
    minimapHeight.value - CURSOR_HEIGHT_OFFSET,
    pos.z,
  );
  spriteMaterial.rotation = rot.y;
}

// --- Custom Render Loop ---
// We hook into the Tres render loop to draw our overlay on top
const onMinimapRender = (ctx: TresContext) => {
  if (!props.show) return;

  const renderer = ctx.renderer.instance;
  const camera = ctx.camera.activeCamera.value;

  renderer.clear();
  renderer.render(sceneStates.value!.tresContext.value!.scene, camera);

  renderer.autoClear = false;
  renderer.render(overlayScene, camera);
  renderer.autoClear = true;
};

watch(
  minimapFrustumSize,
  (frus) => {
    updateCursorSize(frus);
  },
  { immediate: true },
);

watch(minimapHeight, (minimapHeight) => {
  updateCursorDepth(minimapHeight);
});

watch(
  [
    sceneStates.value!.currentCam.value.position,
    sceneStates.value!.currentCam.value.rotation,
  ],
  ([pos, rot]) => {
    updateCursor(pos, rot);
  },
);

const camPos = computed(() => {
  if (mouseOffset.value == undefined) {
    return [
      sceneStates!.value!.currentCam.value!.position.x,
      minimapHeight.value,
      sceneStates!.value!.currentCam.value!.position.z,
    ] as const;
  }
  return [
    mouseOffset.value.x,
    minimapHeight.value,
    mouseOffset.value.y,
  ] as const;
});

function createCursorTexture() {
  const canvas = document.createElement("canvas");
  canvas.width = 128;
  canvas.height = 128;
  const ctx = canvas.getContext("2d")!;
  const centerX = 64,
    centerY = 64;
  const cyan = "#00ffff",
    white = "#ffffff",
    black = "#000000";

  // Beak
  ctx.beginPath();
  ctx.moveTo(centerX - 15, centerY - 25);
  ctx.lineTo(centerX, centerY - 55);
  ctx.lineTo(centerX + 15, centerY - 25);
  ctx.closePath();
  ctx.strokeStyle = black;
  ctx.lineWidth = 6;
  ctx.stroke();
  ctx.fillStyle = cyan;
  ctx.fill();

  // Circle
  ctx.beginPath();
  ctx.arc(centerX, centerY, 30, 0, Math.PI * 2);
  ctx.strokeStyle = black;
  ctx.lineWidth = 10;
  ctx.stroke();
  ctx.strokeStyle = cyan;
  ctx.lineWidth = 4;
  ctx.stroke();

  // Dot
  ctx.beginPath();
  ctx.arc(centerX, centerY, 8, 0, Math.PI * 2);
  ctx.strokeStyle = black;
  ctx.lineWidth = 4;
  ctx.stroke();
  ctx.fillStyle = white;
  ctx.fill();

  return new CanvasTexture(canvas);
}

let foundMove = false;
const mouseOffset = ref<{ x: number; y: number } | undefined>(undefined);

const camera = computed(() => canvas.value?.context?.camera.activeCamera.value);

const canvaDom = computed(
  () => canvas.value?.context?.renderer.instance.domElement,
);

function onPointerDown(_e: PointerEvent) {
  if (canvaDom.value) {
    document.addEventListener("pointermove", onPointerMove);
    document.addEventListener("pointerup", onPointerUp);
  }
  foundMove = false;
}

function onPointerMove(e: PointerEvent) {
  foundMove = true;
  if (mouseOffset.value == undefined) {
    mouseOffset.value = { x: 0, y: 0 };
  }
  mouseOffset.value!.x -=
    (e.movementX * minimapFrustumSize.value) / minimapSize;
  mouseOffset.value!.y -=
    (e.movementY * minimapFrustumSize.value) / minimapSize;
}

function onPointerUp(_e: PointerEvent) {
  if (canvaDom.value) {
    document.removeEventListener("pointermove", onPointerMove);
    document.removeEventListener("pointerup", onPointerUp);
  }
  if (!foundMove) {
    mouseOffset.value = undefined;
  }
}

watch(camera, (camera) => {
  if (camera == undefined) return;
  camera.layers.set(MINIMAP_LAYER);
});

watch(canvaDom, () => {
  canvaDom.value?.addEventListener("pointerdown", onPointerDown);
});

onUnmounted(() => {
  overlayScene.remove();
  cursorSprite.remove();
  spriteMaterial.dispose();
  canvaDom.value?.removeEventListener("pointerdown", onPointerDown);
  document.removeEventListener("pointermove", onPointerMove);
  document.removeEventListener("pointerup", onPointerUp);
});
</script>

<template>
  <div
    v-show="show"
    class="minimap-container absolute bottom-4 right-4 z-20 pointer-events-auto select-none"
    @wheel.prevent="handleMinimapZoom"
  >
    <div class="flex flex-col gap-1 mb-2">
      <label class="text-[15px] text-white uppercase stroked">
        Height: {{ minimapHeight.toFixed(2) }}m
      </label>
      <input
        v-model.number="minimapHeight"
        type="range"
        min="0"
        max="10"
        step="0.01"
        class="slider"
      />
    </div>

    <div
      class="rounded-lg overflow-hidden border border-gray-600/50 shadow-2xl"
    >
      <TresCanvas
        ref="canvas"
        alpha
        clear-color="#111122"
        :window-size="false"
        :width="minimapSize"
        :height="minimapSize"
        :style="{ width: minimapSize + 'px', height: minimapSize + 'px' }"
        @render="onMinimapRender"
      >
        <TresOrthographicCamera
          name="minimapCamera"
          :position="camPos"
          :look-at="[camPos[0], camPos[1] - 1, camPos[2]]"
          :left="-minimapFrustumSize / 2"
          :right="minimapFrustumSize / 2"
          :top="minimapFrustumSize / 2"
          :bottom="-minimapFrustumSize / 2"
          :near="0.1"
          :far="1000"
        />
      </TresCanvas>
    </div>
  </div>
</template>

<style scoped>
.stroked {
  -webkit-text-stroke: 3px #000;
  paint-order: stroke fill;
}
.slider {
  width: 100%;
  accent-color: #00ffff;
}
</style>
