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

const minimapCamera = ref<OrthographicCamera | null>(null);
const isMapOpen = inject(IS_MAP_OPEN_KEY)!;

const aspect = computed(() => {
  const width = sceneStates.currentCam.value.widthRes;
  if (width == 0) return undefined;
  const height = sceneStates.currentCam.value.heightRes || 1;
  return width / height;
});

usePromptUnsaved(sceneStates);

useCameraUpdate(sceneStates);

// ── Raycasting & Input Events (Omitted same logic for brevity) ───────
const raycaster = new Raycaster();
const mouse = new Vector2();

function onCanvasPointer(event: PointerEvent) {
  if (!sceneStates.tresContext.value || !perspectiveCamera.value) return;
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

  // Don't stop to allow changing cubeCamera
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

onMounted(() => {
  stats = new Stats();
  stats.showPanel(0);
  // stats.showPanel(1);
  // stats.showPanel(2); // 0: fps, 1: ms, 2: mb, 3+: custom
  if (document) {
    document.body.appendChild(stats.dom);
  }
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
            v-model="sceneStates.currentCam.value.position[axis]"
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
            v-model="sceneStates.currentCam.value.rotation[axis]"
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
        class="relative flex align-center justify-center items-center"
      >
        <!-- <img
          src="localhost:3000/public/3-real.png"
          class="absolute h-full z-99 opacity-50 pointer-events-none"
        /> -->
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
            :far="50"
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

          <CameraObject
            v-for="[camId, cam] in Object.entries(sceneStates.cameras)"
            :key="camId"
            :cam-id="camId"
            :name="cam.name"
            :workspace="props.workspace"
          />

          <FrustumOverlay />

          <Suspense><Environment preset="city" /></Suspense>
          <TresAmbientLight :intensity="0.02" />
          <TresDirectionalLight :position="[10, 10, 5]" :intensity="0.2" />

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
