<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import {
  tresCanvasParent,
  spectatorCameraPosition,
  spectatorCameraRotation,
  tresContext,
  draggableObjects,
  currentCameraPosition,
  currentCameraRotation,
} from "./refs";
import { SpectatorPosition } from "./spectator-position";
import { SpectatorRotation } from "./spectator-rotation";
import AdjustableInput from "../../adjustable-input/AdjustableInput.vue";
import { SPECTATOR_ADJ_INPUT_SENTIVITY } from "~/constants";
import {
  cameras,
  spawnCameraHere,
  switchToCam,
  switchToSpectator,
} from "./camera-management";
import CameraObject from "../camera-object/CameraObject.vue";
import Scene3dInner from "./Scene3dInner.vue";
import * as THREE from "three";
import type { IUserData } from "./obj-event-handler";

defineProps<{
  modelId?: string | null;
  placeholderText?: string | null;
}>();

onMounted(() => {
  // start loop to move camera from key press
  SpectatorPosition.update();

  // for testing
  if (typeof window !== "undefined") {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).spawnCameraHereNaja = spawnCameraHere;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).switchToCamNaja = switchToCam;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).switchToSpectatorNaja = switchToSpectator;
  }
});

const raycaster = new THREE.Raycaster();
const mouse = new THREE.Vector2();

watch(
  currentCameraPosition,
  (pos) => {
    if (tresContext.value?.camera) {
      tresContext.value.camera.position.set(pos.x, pos.y, pos.z);
    }
  },
  { deep: true },
);

watch(
  currentCameraRotation,
  (pos) => {
    if (tresContext.value?.camera) {
      tresContext.value.camera.rotation.set(pos.x, pos.y, pos.z);
    }
  },
  { deep: true },
);

function onCanvasPointer(event: PointerEvent) {
  const ele = tresContext.value!.renderer.domElement;
  const rect = ele.getBoundingClientRect();

  mouse.x = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height!) * 2 + 1;

  raycaster.setFromCamera(mouse, tresContext.value!.camera!);

  const intersects = raycaster.intersectObjects([...draggableObjects], false);
  if (intersects.length > 0) {
    const foundObj = intersects[0];
    const userData = foundObj?.object.userData as IUserData;
    userData.handleEvent.call(userData, event.type, event);
  } else {
    if (event.type == "pointerdown") {
      SpectatorRotation.onPointerDown(event);
    }
  }
}

watch(
  tresContext,
  (context) => {
    context?.renderer.domElement.addEventListener(
      "pointerdown",
      onCanvasPointer,
    );

    context?.renderer.domElement.addEventListener(
      "pointermove",
      onCanvasPointer,
    );

    context?.renderer.domElement.addEventListener("pointerup", onCanvasPointer);

    context?.renderer.domElement.addEventListener(
      "keydown",
      SpectatorPosition.onKeyDown,
    );

    context?.renderer.domElement.addEventListener(
      "keyup",
      SpectatorPosition.onKeyUp,
    );

    context?.renderer.domElement.addEventListener(
      "blur",
      (event: FocusEvent) => {
        SpectatorRotation.onBlur(event);
        SpectatorPosition.onBlur(event);
      },
    );
  },
  { once: true },
);

// watch(
//   tresCanvasParent,
//   (parent) => {
//     parent?.addEventListener("mousemove", onMouseMove);
//   },
//   { once: true },
// );
</script>

<template>
  <ClientOnly>
    <div ref="tresCanvasParent" class="w-full h-full bg-background relative">
      <div
        id="camera-props"
        class="absolute top-0 right-0 z-10 text-white flex flex-col p-2"
      >
        <p class="text-center w-full h-full">Spectator</p>
        <div class="flex">
          <p>x:</p>
          <AdjustableInput
            v-model="spectatorCameraPosition.x"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>y:</p>
          <AdjustableInput
            v-model="spectatorCameraPosition.y"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>z:</p>
          <AdjustableInput
            v-model="spectatorCameraPosition.z"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>x</sub>:</p>
          <AdjustableInput
            v-model="spectatorCameraRotation.x"
            class="right-adjustable-input"
            :max="Math.PI / 2 - 0.01"
            :min="-Math.PI / 2 + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>y</sub>:</p>
          <AdjustableInput
            v-model="spectatorCameraRotation.y"
            class="right-adjustable-input"
            :max="Math.PI - 0.01"
            :min="-Math.PI + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>z</sub>:</p>
          <AdjustableInput
            v-model="spectatorCameraRotation.z"
            class="right-adjustable-input"
            :max="Math.PI - 0.01"
            :min="-Math.PI + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
      </div>
      <TresCanvas
        id="canvas"
        ref="canvas"
        resize-event="parent"
        clear-color="#0E0C29"
        tabindex="0"
      >
        <!-- Camera -->
        <TresPerspectiveCamera
          ref="spectatorCamera"
          :position="currentCameraPosition"
          :rotation="currentCameraRotation"
          :fov="75"
        />

        <CameraObject
          v-for="[camId, cam] in Object.entries(cameras)"
          :key="camId"
          :name="cam.name"
          :position="cam.position"
          :rotation="cam.rotation"
        />

        <!-- Environment and lighting, from the tresjs/cientos library -->
        <Environment preset="studio" />
        <TresAmbientLight :intensity="0.4" />
        <TresDirectionalLight :position="[10, 10, 5]" :intensity="1" />

        <!-- 3D Objects -->
        <TresMesh ref="meshRef" :position="[0, 0.5, 0]">
          <TresBoxGeometry />
          <TresMeshStandardMaterial
            :color="'#4a90e2'"
            :metalness="0.3"
            :roughness="0.4"
          />
        </TresMesh>

        <!-- Grid -->
        <Grid
          :args="[20, 20]"
          :cell-size="1"
          :cell-thickness="0.5"
          :cell-color="'#4a90e2'"
          :section-size="5"
          :section-thickness="1"
          :section-color="'#ffffff'"
          :fade-distance="50"
          :fade-strength="1"
          infinite-grid
        />
        <Scene3dInner />
      </TresCanvas>
    </div>
  </ClientOnly>
</template>

<style>
#canvas {
  height: 100%;
  width: 100%;
  min-height: 0;
}
#camera-props {
  text-shadow:
    -1px -1px 0 black /* Top-left shadow */,
    1px -1px 0 black /* Top-right shadow */,
    -1px 1px 0 black /* Bottom-left shadow */,
    1px 1px 0 black /* Bottom-right shadow */;
}
</style>

<style scoped>
.right-adjustable-input {
  width: 100%;
  padding: 1px;
  padding-left: 8px;
}
.right-adjustable-input :deep(span) {
  text-align: right;
  width: 100%;
  display: block;
}
.right-adjustable-input :deep(input) {
  text-align: right;
  width: 100%;
}
</style>
