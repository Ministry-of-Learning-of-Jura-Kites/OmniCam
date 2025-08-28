<script setup lang="ts">
const value = defineModel<number>();

const props = defineProps({
  max: {
    type: Number,
    default: null,
  },
  min: {
    type: Number,
    default: null,
  },
  slidingSensitivity: {
    type: Number,
    default: 0.5,
  },
});

const inputEl = ref<HTMLInputElement | null>(null);

const textValue = ref<HTMLSpanElement | null>(null);

const isDragging = ref(false);

const isInputting = ref(false);

let isAdjusting = false;

function onPointerDown(_e: PointerEvent) {
  isDragging.value = true;
  isAdjusting = false;
  isInputting.value = false;
  document.addEventListener("pointerup", onPointerUp);
  setTimeout(async () => {
    if (isDragging.value) {
      isAdjusting = true;
      await textValue?.value?.requestPointerLock();

      document.addEventListener("pointermove", onPointerMove);
    }
  }, 200);
}

function roundTo(num: number, decimals: number) {
  const decimalsPower = Math.pow(10, decimals);
  return Math.round(num * decimalsPower) / decimalsPower;
}

function onPointerUp(_e: PointerEvent) {
  // Is sliding to adjust value
  if (isAdjusting) {
    document.exitPointerLock();
    isDragging.value = false;
    isAdjusting = false;
    document.removeEventListener("pointermove", onPointerMove);
    document.removeEventListener("pointerup", onPointerUp);
    return;
  }
  value.value = roundTo(value.value ?? 0, 5);
  isDragging.value = false;
  isInputting.value = true;

  // Select all on input shown
  watch(
    inputEl,
    (newValue, _oldValue) => {
      newValue?.select();
    },
    { once: true },
  );

  // Listen to click outside of input to change isInputting
  document.addEventListener("click", handleClickOutside);

  document.removeEventListener("pointerup", onPointerUp);
}

function onPointerMove(e: MouseEvent) {
  if (!isDragging.value) {
    return;
  }
  const deltaX = e.movementX;
  if (deltaX > 200) {
    return;
  }
  if (value?.value != null) {
    const newVal = value.value + deltaX * props.slidingSensitivity;
    setClamp(newVal);
  }
}

function setClamp(input: number) {
  let newVal = input;
  if (props.max != null) {
    newVal = Math.min(props.max, newVal);
  }
  if (props.min != null) {
    newVal = Math.max(props.min, newVal);
  }
  value.value = newVal;
}

function unsetInputting() {
  isInputting.value = false;
  document.removeEventListener("click", handleClickOutside);
}

function handleClickOutside(e: MouseEvent) {
  if (inputEl.value && !inputEl.value.contains(e.target as Node)) {
    unsetInputting();
  }
}
</script>

<template>
  <div>
    <span
      v-if="!isInputting"
      ref="textValue"
      class="text-right w-full select-none"
      @pointerdown="onPointerDown"
      >{{ (value ?? 0).toFixed(2) }}</span
    >
    <input
      v-if="isInputting"
      ref="inputEl"
      v-model="value"
      type="number"
      :size="value?.toString().length"
      @focusout="unsetInputting"
      @keypress.enter="unsetInputting"
    />
  </div>
</template>

<style>
input {
  field-sizing: content;
}

/* For WebKit browsers (Chrome, Safari) */
input[type="number"]::-webkit-outer-spin-button,
input[type="number"]::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0; /* Important for removing extra space */
}

/* For Mozilla Firefox */
input[type="number"] {
  appearance: textfield;
  -moz-appearance: textfield;
}
input {
  width: 100%;
  border-radius: 5px;
  border: 1px solid black;
  outline: 1px solid white;
  box-sizing: border-box;
}
</style>
