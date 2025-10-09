<script setup lang="ts">
// import type * as THREE from "three";
import { SCENE_STATES_KEY } from "../scene-states-provider/create-scene-states";
import MovableArrow from "../movable-arrow/MovableArrow.vue";
import TresMesh from "@tresjs/core";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";
import CameraFrustum from "../camera-frustum/CameraFrustum.vue";

const props = defineProps({
  name: {
    type: String,
    default: "Untitled",
  },
  camId: {
    type: String,
    default: "",
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const cam = toRef(sceneStates.cameras, props.camId);
</script>

<template>
  <TresMesh :position="sceneStates.cameras[props.camId]!.position.clone()">
    <TresMesh :rotation="sceneStates.cameras[props.camId]!.rotation.clone()">
      <TresBoxGeometry :args="[0.5, 0.5, 0.5]" />
      <TresMeshBasicMaterial color="white" />
      <CameraFrustum :fov="sceneStates.cameras[props.camId]!.fov" />
    </TresMesh>
    <template v-if="cam != null">
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isHidingArrows || sceneStates.currentCamId.value == props.camId
        "
        :controlling="cam.controlling"
        direction="x"
        color="green"
        @change="sceneStates.markedForCheck.add(camId)"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isHidingArrows || sceneStates.currentCamId.value == props.camId
        "
        :controlling="cam.controlling"
        direction="y"
        color="red"
        @change="sceneStates.markedForCheck.add(camId)"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isHidingArrows || sceneStates.currentCamId.value == props.camId
        "
        :controlling="cam.controlling"
        direction="z"
        color="blue"
        @change="sceneStates.markedForCheck.add(camId)"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isHidingWheels || sceneStates.currentCamId.value == props.camId
        "
        direction="x"
        color="green"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isHidingWheels || sceneStates.currentCamId.value == props.camId
        "
        direction="y"
        color="red"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isHidingWheels || sceneStates.currentCamId.value == props.camId
        "
        direction="z"
        color="blue"
      />
    </template>
  </TresMesh>
</template>
