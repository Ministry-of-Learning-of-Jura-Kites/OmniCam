<script setup lang="ts">
import { GridHelper, Vector3 } from "three";
import { IS_CALIBRATING_KEY } from "~/constants/state-keys";
import type { MovableObject } from "~/types/movable";

import MovableArrow from "../movable-arrow/MovableArrow.vue";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const isCalibrating = inject(IS_CALIBRATING_KEY);

const movableObject = shallowReactive<MovableObject>({
  position: new Vector3(0, 0, 0),
  controlling: undefined,
});

const triggerUpdate = () => {
  movableObject.position = movableObject.position.clone();
};

// v-if="isCalibrating" && props.workspace != null"
</script>

<template>
  <TresGroup
    v-if="isCalibrating && props.workspace != null"
    :position-x="movableObject.position.x"
    :position-y="movableObject.position.y"
    :position-z="movableObject.position.z"
  >
    <primitive
      ref="gridRef"
      :object="new GridHelper(1, 10, 0x00ff66, 0xdb6060)"
    />

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
  </TresGroup>
</template>
