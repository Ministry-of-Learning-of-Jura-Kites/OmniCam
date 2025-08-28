<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import { useSpectatorPosition } from "./use-spectator-position";
import { useSpectatorRotation } from "./use-spectator-rotation";
import AdjustableInput from "../../adjustable-input/AdjustableInput.vue";
import { SPECTATOR_ADJ_INPUT_SENTIVITY } from "~/constants";
import { useCameraManagement } from "./use-camera-management";
import CameraObject from "../camera-object/CameraObject.vue";
import Scene3dInner from "./inner/Scene3dInner.vue";
import * as THREE from "three";
import type { IUserData } from "./obj-event-handler";
import { createSceneStates, SCENE_STATES_KEY } from "./use-scene-state";
import { useCameraUpdate } from "./use-camera-update";

defineProps<{
  modelId?: string | null;
  placeholderText?: string | null;
}>();

const sceneStates = createSceneStates();

provide(SCENE_STATES_KEY, sceneStates);

const spectatorPosition = useSpectatorPosition(sceneStates);
const spectatorRotation = useSpectatorRotation(sceneStates);

const cameraManagement = useCameraManagement(sceneStates);

useCameraUpdate(sceneStates);

onMounted(() => {
  // start loop to move camera from key press
  spectatorPosition.update();

  // for testing
  if (typeof window !== "undefined") {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).spawnCameraHereNaja = cameraManagement.spawnCameraHere;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).switchToCamNaja = cameraManagement.switchToCam;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).switchToSpectatorNaja = cameraManagement.switchToSpectator;
  }
});

const raycaster = new THREE.Raycaster();
const mouse = new THREE.Vector2();

function onCanvasPointer(event: PointerEvent) {
  const ele = sceneStates.tresContext.value!.renderer.domElement;
  const rect = ele.getBoundingClientRect();

  mouse.x = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height!) * 2 + 1;

  raycaster.setFromCamera(mouse, sceneStates.tresContext.value!.camera!);

  const intersects = raycaster.intersectObjects(
    [...sceneStates.draggableObjects],
    false,
  );
  if (intersects.length > 0) {
    const foundObj = intersects[0];
    const userData = foundObj?.object.userData as IUserData;
    userData.handleEvent.call(userData, event.type, event);
  } else {
    if (event.type == "pointerdown") {
      spectatorRotation.onPointerDown(event);
    }
  }
}

watch(
  sceneStates.tresContext,
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
      spectatorPosition.onKeyDown,
    );

    context?.renderer.domElement.addEventListener(
      "keyup",
      spectatorPosition.onKeyUp,
    );

    context?.renderer.domElement.addEventListener(
      "blur",
      (event: FocusEvent) => {
        spectatorRotation.onBlur(event);
        spectatorPosition.onBlur(event);
      },
    );
  },
  { once: true },
);
</script>

<template>
  <ClientOnly>
    <div
      :ref="sceneStates.tresCanvasParent"
      class="w-full h-full bg-background relative"
    >
      <div
        id="camera-props"
        class="absolute top-0 right-0 z-10 text-white flex flex-col p-2"
      >
        <p class="text-center w-full h-full">Spectator</p>
        <div class="flex">
          <p>x:</p>
          <AdjustableInput
            v-model="sceneStates.spectatorCameraPosition.x"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>y:</p>
          <AdjustableInput
            v-model="sceneStates.spectatorCameraPosition.y"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>z:</p>
          <AdjustableInput
            v-model="sceneStates.spectatorCameraPosition.z"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>x</sub>:</p>
          <AdjustableInput
            v-model="sceneStates.spectatorCameraRotation.x"
            class="right-adjustable-input"
            :max="Math.PI / 2 - 0.01"
            :min="-Math.PI / 2 + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>y</sub>:</p>
          <AdjustableInput
            v-model="sceneStates.spectatorCameraRotation.y"
            class="right-adjustable-input"
            :max="Math.PI - 0.01"
            :min="-Math.PI + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>z</sub>:</p>
          <AdjustableInput
            v-model="sceneStates.spectatorCameraRotation.z"
            class="right-adjustable-input"
            :max="Math.PI - 0.01"
            :min="-Math.PI + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
      </div>
      <TresCanvas
        id="canvas"
        resize-event="parent"
        clear-color="#0E0C29"
        tabindex="0"
      >
        <!-- Camera -->
        <TresPerspectiveCamera
          :position="sceneStates.currentCameraPosition.value"
          :rotation="sceneStates.currentCameraRotation.value"
          :fov="75"
        />

        <CameraObject
          v-for="[camId, cam] in Object.entries(sceneStates.cameras)"
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
        <TresMesh :position="[0, 0.5, 0]">
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
