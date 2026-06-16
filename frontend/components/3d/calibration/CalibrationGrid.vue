<script setup lang="ts">
import { Euler, GridHelper, Vector3, Quaternion } from "three";
import { PANEL_KEY } from "~/constants/state-keys";
import type { MovableObject } from "~/types/movable";
import MovableArrow from "../movable-arrow/MovableArrow.vue";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";
const props = withDefaults(
  defineProps<{
    initialPos?: [number, number, number];
    workspace?: string | null;
  }>(),
  { initialPos: () => [0, 0, 0] as const, workspace: null },
);
const { currentPanel, calibrationPanelInfo } = inject(PANEL_KEY)!;
const { calibrationGridScale } = calibrationPanelInfo;
const movableObject = reactive<MovableObject>({
  position: new Vector3(
    props.initialPos[0],
    props.initialPos[1],
    props.initialPos[2],
  ),
  rotation: new Euler(0, 0, 0),
  controlling: undefined,
});
const gridHelper = markRaw(new GridHelper(1, 10, 0x00ff66, 0xdb6060));

const triggerUpdate = () => {
  movableObject.position = movableObject.position.clone();
  movableObject.rotation = movableObject.rotation.clone();
};
const rotationQuat = computed(() => {
  const quaternion = new Quaternion().setFromEuler(movableObject!.rotation);
  return quaternion;
});

onUnmounted(() => {
  gridHelper.geometry.dispose();

  if (Array.isArray(gridHelper.material)) {
    gridHelper.material.forEach((m) => m.dispose());
  } else {
    gridHelper.material.dispose();
  }
});
</script>

<template>
  <TresGroup
    v-if="currentPanel == 'calibration' && props.workspace == 'me'"
    :position-x="movableObject.position.x"
    :position-y="movableObject.position.y"
    :position-z="movableObject.position.z"
  >
    <TresObject3D :quaternion="rotationQuat">
      <primitive
        ref="gridRef"
        :object="gridHelper"
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
      @move="triggerUpdate"
    />
    <MovableArrow
      v-model="movableObject"
      :controlling="movableObject.controlling"
      direction="z"
      color="blue"
      @move="triggerUpdate"
    />
    <RotationWheel
      v-model="movableObject"
      :controlling="movableObject.controlling"
      direction="y"
      color="red"
      @move="triggerUpdate"
    />
  </TresGroup>
</template>
