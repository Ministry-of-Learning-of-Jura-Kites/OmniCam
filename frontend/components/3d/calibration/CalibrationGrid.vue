<script setup lang="ts">
import { GridHelper } from "three";
import { IS_CALIBRATING_KEY } from "~/constants/state-keys";

import MovableArrow from "../movable-arrow/MovableArrow.vue";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const isCalibrating = inject(IS_CALIBRATING_KEY);
</script>

<template v-if="isCalibrating && props.workspace != null">
  <TresGroup>
    <primitive
      ref="gridRef"
      :object="new GridHelper(1, 10, 0x00ff66, 0xdb6060)"
      :position="[0, 0, 0]"
    />
  </TresGroup>

  <template v-if="true">
    <!-- <MovableArrow
      v-model=""
      direction="x"
      color="green"
    />
    <MovableArrow
      direction="y"
      color="red"
    />
    <MovableArrow
      direction="z"
      color="blue"
    /> -->
  </template>
</template>
