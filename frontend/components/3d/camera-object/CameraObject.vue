<script setup lang="ts">
import type * as THREE from "three";
import { SCENE_STATES_KEY } from "../scene-states-provider/create-scene-states";
import MovableArrow from "../movable-arrow/MovableArrow.vue";

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

const cameraMesh = ref<THREE.Mesh | null>(null);

// watch(cameraMesh, (mesh) => {
//   console.log("gggg", mesh);
//   mesh!.position.copy(props.position.value);
// });

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
  <TresMesh
    ref="cameraMesh"
    :position="sceneStates.cameras[props.camId]!.position.clone()"
  >
    <TresMesh :rotation="sceneStates.cameras[props.camId]!.rotation.clone()">
      <TresBoxGeometry :args="[0.5, 0.5, 0.5]" />
      <TresMeshBasicMaterial color="white" />
    </TresMesh>
    <MovableArrow
      v-if="cameraMesh != null"
      direction="x"
      :camera-mesh="cameraMesh"
      color="green"
    />
    <MovableArrow
      v-if="cameraMesh != null"
      direction="y"
      :camera-mesh="cameraMesh"
      color="red"
    />
    <MovableArrow
      v-if="cameraMesh != null"
      direction="z"
      :camera-mesh="cameraMesh"
      color="blue"
    />
  </TresMesh>
</template>
