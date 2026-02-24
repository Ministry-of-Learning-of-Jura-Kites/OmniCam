<script setup lang="ts">
import { Euler, GridHelper, Vector3, Quaternion } from "three";
import {
  IS_CALIBRATING_KEY,
  CALIBRATION_GRID_SCALE,
} from "~/constants/state-keys";
import type { MovableObject } from "~/types/movable";
import MovableArrow from "../movable-arrow/MovableArrow.vue";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";
const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});
const isCalibrating = inject(IS_CALIBRATING_KEY);
const calibrationGridScale = inject(CALIBRATION_GRID_SCALE);
const movableObject = reactive<MovableObject>({
  position: new Vector3(0, 0, 0),
  rotation: new Euler(0, 0, 0),
  controlling: undefined,
});
const triggerUpdate = () => {
  movableObject.position = movableObject.position.clone();
  movableObject.rotation = movableObject.rotation.clone();
};
const rotationQuat = computed(() => {
  const quaternion = new Quaternion().setFromEuler(movableObject!.rotation);
  return quaternion;
});
// v-if="isCalibrating" && props.workspace != null"
</script>

<template>
  <TresGroup
    v-if="isCalibrating && props.workspace != null"
    :position-x="movableObject.position.x"
    :position-y="movableObject.position.y"
    :position-z="movableObject.position.z"
  >
    <TresObject3D :quaternion="rotationQuat">
      <primitive
        ref="gridRef"
        :object="new GridHelper(1, 10, 0x00ff66, 0xdb6060)"
        :scale="[calibrationGridScale, 1, calibrationGridScale]"
      />
    </TresObject3D>

    <MovableArrow
      v-model="movableObject"
      :controlling="movableObject.controlling"
      direction="x"
      color="green"
      @change="triggerUpdate"
    />
    <MovableArrow
      v-model="movableObject"
      :controlling="movableObject.controlling"
      direction="y"
      color="red"
      @change="triggerUpdate"
    />
    <MovableArrow
      v-model="movableObject"
      :controlling="movableObject.controlling"
      direction="z"
      color="blue"
      @change="triggerUpdate"
    />
    <RotationWheel
      v-model="movableObject"
      :controlling="movableObject.controlling"
      direction="y"
      color="red"
      @change="triggerUpdate"
    />
  </TresGroup>
</template>
