<script setup lang="ts">
import { watchDebounced } from "@vueuse/core";
import { Navigation2 } from "lucide-vue-next";
import { Vector3 } from "three";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const props = defineProps<{
  targetPos: Vector3;
  label?: string;
  show: boolean;
}>();

const sceneStates = inject(SCENE_STATES_KEY)!;

const camera = computed(() => {
  return sceneStates.value!.perspectiveCamera.value;
});

const isOffscreen = ref(false);

const style = reactive({
  left: "0px",
  top: "0px",
  arrowRotation: "0deg",
});

const projectedPos = new Vector3();
const cameraForward = new Vector3();
const cameraRight = new Vector3();
const cameraUp = new Vector3();
const targetDirection = new Vector3();

watch(
  () => [
    sceneStates.value!.currentCam.value.position,
    sceneStates.value!.currentCam.value.rotation,
    props.targetPos,
    sceneStates.value!.screenSize,
    props.show,
  ],
  () => {
    if (!props.show) return;

    const width = sceneStates.value!.screenSize.width!;
    const height = sceneStates.value!.screenSize.height!;
    const cam = camera.value!;

    const pos = projectedPos.copy(props.targetPos).project(cam);
    const isBehind = pos.z > 1;

    const inside =
      !isBehind && pos.x >= -1 && pos.x <= 1 && pos.y >= -1 && pos.y <= 1;

    isOffscreen.value = !inside;

    if (!isOffscreen.value) {
      style.left = `${(pos.x * 0.5 + 0.5) * width}px`;
      style.top = `${(-pos.y * 0.5 + 0.5) * height}px`;
    }
  },
  { deep: true, immediate: true },
);

watchDebounced(
  () => [
    sceneStates.value!.currentCam.value.position,
    sceneStates.value!.currentCam.value.rotation,
    props.targetPos,
    sceneStates.value!.screenSize,
    props.show,
  ],
  () => {
    if (!props.show || !isOffscreen.value) return;

    const width = sceneStates.value!.screenSize.width!;
    const height = sceneStates.value!.screenSize.height!;
    const cam = camera.value!;

    const pos = projectedPos.copy(props.targetPos).project(cam);
    const isBehind = pos.z > 1;
    const margin = 80;

    if (!isBehind) {
      const t = 1 / Math.max(Math.abs(pos.x), Math.abs(pos.y));
      const edgeX = pos.x * t;
      const edgeY = pos.y * t;

      style.left = `${width / 2 + edgeX * (width / 2 - margin)}px`;
      style.top = `${height / 2 - edgeY * (height / 2 - margin)}px`;

      const angle = Math.atan2(pos.x, pos.y);
      style.arrowRotation = `${(angle * 180) / Math.PI}deg`;
      return;
    }

    targetDirection.copy(props.targetPos).sub(cam.position).normalize();
    cam.getWorldDirection(cameraForward);
    cameraRight.crossVectors(cameraForward, cam.up).normalize();
    cameraUp.copy(cam.up).normalize();

    const horizontal = targetDirection.dot(cameraRight);
    const vertical = targetDirection.dot(cameraUp);
    const absH = Math.abs(horizontal);
    const absV = Math.abs(vertical);

    const useVertical = absV > absH * 1.35;

    if (!useVertical) {
      const isLeft = horizontal < 0;
      style.left = `${isLeft ? margin : width - margin}px`;
      style.top = `${height / 2}px`;
      style.arrowRotation = isLeft ? "-90deg" : "90deg";
    } else {
      const isUp = vertical > 0;
      style.left = `${width / 2}px`;
      style.top = `${isUp ? margin : height - margin}px`;
      style.arrowRotation = isUp ? "0deg" : "180deg";
    }
  },
  { deep: true, immediate: true, debounce: 10, maxWait: 50 },
);
</script>

<template>
  <div
    v-if="props.show"
    class="beacon-container"
    :style="{
      left: style.left,
      top: style.top,
    }"
  >
    <template v-if="!isOffscreen">
      <div class="dot">
        <span class="label">
          {{ label }}
        </span>
      </div>
    </template>

    <template v-else>
      <div class="offscreen-wrapper">
        <Navigation2
          class="arrow"
          :style="{
            transform: `rotate(${style.arrowRotation})`,
          }"
        />

        <span class="offscreen-label">
          {{ label }}
        </span>
      </div>
    </template>
  </div>
</template>

<style scoped>
.beacon-container {
  position: absolute;
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 9;
}

.dot {
  width: 12px;
  height: 12px;
  background: #00ffcc;
  border: 2px solid white;
  border-radius: 999px;
  box-shadow: 0 0 15px #00ffcc;
  position: relative;
}

.label {
  position: absolute;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  white-space: nowrap;
  background: rgba(0, 0, 0, 0.75);
  color: white;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
}

.offscreen-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.75);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 999px;
  backdrop-filter: blur(6px);
}

.arrow {
  width: 18px;
  height: 18px;
  color: #00ffcc;
  flex-shrink: 0;
  transition: transform 0.15s ease;
}

.offscreen-label {
  color: white;
  font-size: 12px;
  white-space: nowrap;
}
</style>
