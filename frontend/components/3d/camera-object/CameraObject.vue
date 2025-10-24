<script setup lang="ts">
// import type * as THREE from "three";
import { SCENE_STATES_KEY } from "../scene-states-provider/create-scene-states";
import MovableArrow from "../movable-arrow/MovableArrow.vue";
import TresMesh from "@tresjs/core";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";
import CameraFrustum from "../camera-frustum/CameraFrustum.vue";
import * as THREE from "three";

const props = defineProps({
  name: {
    type: String,
    default: "Untitled",
  },
  camId: {
    type: String,
    default: "",
  },
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const cam = toRef(sceneStates.cameras, props.camId);

const camQuat = computed(() => {
  const quaternion = new THREE.Quaternion().setFromEuler(cam!.value.rotation);
  return quaternion;
});
</script>

<template>
  <TresMesh :position="[cam!.position.x, cam!.position.y, cam!.position.z]">
    <TresObject3D :quaternion="camQuat">
      <!-- Use quaternion for applying on top of local rotation -->
      <TresMesh :rotation="[Math.PI / 2, 0, 0]" :position="[0, 0, 0.25 + 0.06]">
        <TresCylinderGeometry :args="[0.2, 0.2, 0.5]" />
        <TresMeshBasicMaterial color="white" />
      </TresMesh>
      <TresMesh :rotation="[Math.PI / 2, 0, 0]" :position="[0, 0, 0.06]">
        <TresCylinderGeometry :args="[0.05, 0.05, 0.12]" />
        <TresMeshBasicMaterial color="black" />
      </TresMesh>
      <TresMesh :rotation="[Math.PI / 2, 0, 0]" :position="[0, 0, 0.5 + 0.1]">
        <TresCylinderGeometry :args="[0.05, 0.05, 0.1]" />
        <TresMeshBasicMaterial color="white" />
      </TresMesh>
      <TresMesh :rotation="[Math.PI / 5, 0, 0]" :position="[0, 0.07, 0.71]">
        <TresCylinderGeometry :args="[0.05, 0.05, 0.25]" />
        <TresMeshBasicMaterial color="white" />
      </TresMesh>
      <TresMesh :rotation="[Math.PI / 5, 0, 0]" :position="[0, 0.17, 0.78]">
        <TresCylinderGeometry :args="[0.15, 0.15, 0.02]" />
        <TresMeshBasicMaterial color="white" />
      </TresMesh>
      <CameraFrustum :fov="cam!.fov" :is-hiding="cam!.isHidingFrustum" />
    </TresObject3D>
    <template v-if="cam != null">
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isHidingArrows ||
          sceneStates.currentCamId.value == props.camId ||
          props.workspace == null
        "
        :controlling="cam.controlling"
        direction="x"
        color="green"
        @change="sceneStates.markedForCheck.add(camId)"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isHidingArrows ||
          sceneStates.currentCamId.value == props.camId ||
          props.workspace == null
        "
        :controlling="cam.controlling"
        direction="y"
        color="red"
        @change="sceneStates.markedForCheck.add(camId)"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isHidingArrows ||
          sceneStates.currentCamId.value == props.camId ||
          props.workspace == null
        "
        :controlling="cam.controlling"
        direction="z"
        color="blue"
        @change="sceneStates.markedForCheck.add(camId)"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isHidingWheels ||
          sceneStates.currentCamId.value == props.camId ||
          props.workspace == null
        "
        direction="x"
        color="green"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isHidingWheels ||
          sceneStates.currentCamId.value == props.camId ||
          props.workspace == null
        "
        direction="y"
        color="red"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isHidingWheels ||
          sceneStates.currentCamId.value == props.camId ||
          props.workspace == null
        "
        direction="z"
        color="blue"
      />
    </template>
  </TresMesh>
</template>
