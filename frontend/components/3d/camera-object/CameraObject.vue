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
    <MovableArrow
      v-if="cam != null"
      v-model="cam"
      direction="x"
      color="green"
    />
    <MovableArrow v-if="cam != null" v-model="cam" direction="y" color="red" />
    <MovableArrow v-if="cam != null" v-model="cam" direction="z" color="blue" />
    <RotationWheel v-if="cam != null" v-model="cam" type="x" color="green" />
    <RotationWheel v-if="cam != null" v-model="cam" type="y" color="red" />
    <RotationWheel v-if="cam != null" v-model="cam" type="z" color="blue" />
  </TresMesh>
</template>
