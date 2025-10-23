<script setup lang="ts">
import { TresCanvas } from "@tresjs/core";
import { Grid, Environment } from "@tresjs/cientos";
import AdjustableInput from "../../adjustable-input/AdjustableInput.vue";
import { SPECTATOR_ADJ_INPUT_SENTIVITY } from "~/constants";
import CameraObject from "../camera-object/CameraObject.vue";
import * as THREE from "three";
import { SCENE_STATES_KEY } from "~/components/3d/scene-states-provider/create-scene-states";
import { useCameraUpdate } from "./use-camera-update";
import type { IUserData } from "~/types/obj-3d-user-data";
import ModelLoader from "../model-loader/ModelLoader.vue";

const _ = defineProps({
  projectId: {
    type: String,
    required: true,
  },
  modelId: {
    type: String,
    required: true,
  },
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const { data: modelResp } = sceneStates.modelInfo;

const camera: Ref<THREE.PerspectiveCamera | null> = ref(null);

const canvas: Ref<InstanceType<typeof TresCanvas> | null> = ref(null);

useCameraUpdate(sceneStates);

onMounted(() => {
  // start loop to move camera from key press
  sceneStates.spectatorPosition.refreshCameraState();
});

const raycaster = new THREE.Raycaster();
const mouse = new THREE.Vector2();

function onCanvasPointer(event: PointerEvent) {
  const ele = sceneStates.tresContext.value!.renderer.domElement;
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
  } else {
    if (event.type == "pointerdown") {
      sceneStates.spectatorRotation.onPointerDown(event);
    }
  }
}

watch(
  canvas,
  (canvas) => {
    const context = canvas!.context!;
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

    renderer.value.domElement.addEventListener("blur", (event: FocusEvent) => {
      sceneStates.spectatorRotation.onBlur(event);
      sceneStates.spectatorPosition.onBlur(event);
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
    const newFov = transform?.value?.fov ?? cam!.value!.fov;
    if (camera.value) {
      camera.value.fov = newFov!;
      camera.value.updateProjectionMatrix();
    }
  },
  { deep: true },
);

// onMounted(() => {
//   useAutosave(sceneStates, props.workspace);
// });
</script>

<template>
  <ClientOnly>
    <div
      :ref="sceneStates.tresCanvasParent"
      class="w-full h-full bg-background relative"
    >
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
          class="bg-black align-start pointer-events-auto"
        ></div>
        <div
          :style="{
            width: sceneStates.aspectMargin.width ?? '0px',
            height: sceneStates.aspectMargin.height ?? '0px',
          }"
          class="bg-black align-end pointer-events-auto"
        ></div>
      </div>
      <div
        id="camera-props"
        class="absolute top-0 right-0 z-10 text-white flex flex-col p-2"
      >
        <p class="text-center w-full h-full">Spectator</p>
        <div class="flex">
          <p>x:</p>
          <AdjustableInput
            v-model="spectatorRefs.position.x"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>y:</p>
          <AdjustableInput
            v-model="spectatorRefs.position.y"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>z:</p>
          <AdjustableInput
            v-model="spectatorRefs.position.z"
            class="right-adjustable-input"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>x</sub>:</p>
          <AdjustableInput
            v-model="spectatorRefs.rotation.x"
            class="right-adjustable-input"
            :max="Math.PI / 2 - 0.01"
            :min="-Math.PI / 2 + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>y</sub>:</p>
          <AdjustableInput
            v-model="spectatorRefs.rotation.y"
            class="right-adjustable-input"
            :max="Math.PI - 0.01"
            :min="-Math.PI + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>θ<sub>z</sub>:</p>
          <AdjustableInput
            v-model="spectatorRefs.rotation.z"
            class="right-adjustable-input"
            :max="Math.PI - 0.01"
            :min="-Math.PI + 0.01"
            :sliding-sensitivity="SPECTATOR_ADJ_INPUT_SENTIVITY"
          />
        </div>
        <div class="flex">
          <p>VFOV:</p>
          <AdjustableInput
            v-model="sceneStates.spectatorCameraFov"
            class="right-adjustable-input"
            :max="180"
            :min="0"
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
        />

        <CameraObject
          v-for="[camId, cam] in Object.entries(sceneStates.cameras)"
          :key="camId"
          :cam-id="camId"
          :name="cam.name"
        />

        <!-- Environment and lighting, from the tresjs/cientos library -->
        <Suspense>
          <Environment preset="city" />
        </Suspense>
        <TresAmbientLight :intensity="0.4" />
        <TresDirectionalLight :position="[10, 10, 5]" :intensity="1" />

        <!-- 3D Objects -->
        <Suspense>
          <ModelLoader :path="modelResp.filePath" :position="[0, 0, 0]" />
        </Suspense>

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
          :infinite-grid="true"
          :side="THREE.DoubleSide"
        />
        <!-- <Scene3dInner /> -->
      </TresCanvas>
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
