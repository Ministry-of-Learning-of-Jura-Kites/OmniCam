<script setup lang="ts">
import { SCENE_STATES_KEY } from "@/constants/state-keys";

import MovableArrow from "../movable-arrow/MovableArrow.vue";
import TresMesh from "@tresjs/core";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";
import CameraFrustum from "../camera-frustum/CameraFrustum.vue";
import { Quaternion } from "three";
import { safeGetAspectRatio } from "~/utils/aspect-ratio";

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
  const quaternion = new Quaternion().setFromEuler(cam!.value.rotation);
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
      <CameraFrustum
        :id="camId"
        :fov="cam!.fov"
        :aspect="safeGetAspectRatio(cam.aspectWidth, cam.aspectHeight)"
        :length="cam!.frustumLength"
        :color="cam!.frustumColor"
        :is-hiding="
          cam!.isHidingFrustum || camId == sceneStates.currentCamId.value
        "
      />
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
