<script setup lang="ts">
// import type * as THREE from "three";
import { SCENE_STATES_KEY } from "../scene-states-provider/create-scene-states";
import MovableArrow from "../movable-arrow/MovableArrow.vue";
import TresMesh from "@tresjs/core";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";

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

// const cameraMesh = useTemplateRef<THREE.Mesh>("cameraMesh");

const cam = toRef(sceneStates.cameras, props.camId);

// watch(
//   props.position,
//   (pos) => {
//     if (cameraMesh.value != null) {
//       console.log("");
//       cameraMesh.value!.position.set(pos[0], pos[1], pos[2]);
//     }
//   },
//   { deep: true, immediate: true },
// );
</script>

<template>
  <TresMesh :position="sceneStates.cameras[props.camId]!.position.clone()">
    <TresMesh :rotation="sceneStates.cameras[props.camId]!.rotation.clone()">
      <TresBoxGeometry :args="[0.5, 0.5, 0.5]" />
      <TresMeshBasicMaterial color="white" />
    </TresMesh>
    <template v-if="cam != null">
      <MovableArrow
        v-model="cam"
        :is-hiding="cam.isHidingArrows"
        direction="x"
        color="green"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="cam.isHidingArrows"
        direction="y"
        color="red"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="cam.isHidingArrows"
        direction="z"
        color="blue"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="cam.isHidingWheels"
        type="x"
        color="green"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="cam.isHidingWheels"
        type="y"
        color="red"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="cam.isHidingWheels"
        type="z"
        color="blue"
      />
    </template>
  </TresMesh>
</template>
