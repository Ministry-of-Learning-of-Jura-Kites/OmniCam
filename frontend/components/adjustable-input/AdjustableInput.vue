<script setup lang="ts">
import { cameraPosition } from "../scene-3d/refs";
import type { Reactive } from "vue";

const value = defineModel<number>();

const isDragging = ref(false);

const textValue: Ref<HTMLSpanElement | null> = ref(null);

async function onPointerDown(e: PointerEvent) {
  isDragging.value = true;
  await textValue?.value?.requestPointerLock();

  document.addEventListener("pointermove", onPointerMove);
  document.addEventListener("pointerup", onPointerUp);
}

function onPointerUp(e: PointerEvent) {
  document.exitPointerLock();
  isDragging.value = false;
  document.removeEventListener("pointermove", onPointerMove);
  document.removeEventListener("pointerup", onPointerUp);
}

function onPointerMove(e: MouseEvent) {
  if (!isDragging.value) {
    return;
  }
  const deltaX = e.movementX;
  if (value?.value != null) {
    value.value += deltaX / 20;
  }
}
</script>

<template>
  <span
    ref="textValue"
    class="text-right w-full select-none"
    @pointerdown="onPointerDown"
    >{{ Math.round((value ?? 0) * 100) / 100 }}</span
  >
</template>
