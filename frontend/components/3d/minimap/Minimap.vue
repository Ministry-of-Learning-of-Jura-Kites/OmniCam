<script setup lang="ts">
import { type OrthographicCamera, WebGLRenderer, type Scene } from "three";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const { minimapCamera, show } = defineProps({
  minimapCamera: {
    type: Object as PropType<OrthographicCamera | null>,
    default: null,
  },
  show: {
    type: Boolean,
    default: true,
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const minimapCanvas = ref<HTMLCanvasElement | null>(null);
let miniRenderer: WebGLRenderer | null = null;
let miniCamUpdate = false;
// eslint-disable-next-line @typescript-eslint/no-unused-vars
let rafId: number | null = null;

// UPDATED: Starting height for the "slicing" view
const minimapHeight: Ref<number> = ref(5);

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

function render() {
  if (
    !minimapCanvas.value ||
    !minimapCamera ||
    !sceneStates.tresContext.value ||
    !show
  ) {
    rafId = requestAnimationFrame(render);
    return;
  }

  const mainScene = sceneStates.tresContext.value.scene as Scene;
  const miniCam = minimapCamera as OrthographicCamera;

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
      sceneStates.spectatorCameraPosition.x,
      minimapHeight.value,
      sceneStates.spectatorCameraPosition.z,
    );

    // Look straight down from that height
    miniCam.lookAt(
      sceneStates.spectatorCameraPosition.x,
      minimapHeight.value - 1,
      sceneStates.spectatorCameraPosition.z,
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
  rafId = requestAnimationFrame(render);
}

onMounted(() => {
  rafId = requestAnimationFrame(render);
});

onUnmounted(() => {
  if (miniRenderer) {
    miniRenderer.dispose();
    miniRenderer.forceContextLoss(); // Help the browser release the WebGL context
    miniRenderer = null;
  }
});

watch(
  [minimapFrustumSize, minimapHeight, sceneStates.spectatorCameraPosition],
  () => {
    miniCamUpdate = true;
  },
  { immediate: true },
);
</script>

<template>
  <div
    v-show="show"
    class="minimap-container absolute bottom-4 right-4 z-20 pointer-events-auto select-none"
    @wheel.prevent="handleMinimapZoom"
  >
    <div class="flex flex-col gap-1 mb-2">
      <label class="text-[10px] text-white opacity-70 uppercase stroked">
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
      class="minimap-label text-white text-[10px] text-center mb-1 opacity-80 uppercase tracking-widest stroked"
    >
      Minimap (Scroll to Zoom)
    </div>
    <canvas
      ref="minimapCanvas"
      class="rounded-lg shadow-2xl border border-gray-600/50 cursor-zoom-in bg-black"
    ></canvas>
  </div>
</template>

<style lang="css" scoped>
.stroked {
  -webkit-text-stroke: 2px #000;
  paint-order: stroke fill;
  color: white;
}
</style>
