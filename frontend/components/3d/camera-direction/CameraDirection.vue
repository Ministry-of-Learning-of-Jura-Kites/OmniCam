<script setup lang="ts">
import { watchDebounced } from "@vueuse/core";
import { Navigation2 } from "lucide-vue-next";
import { Vector3 } from "three";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const props = defineProps<{
  targetPos: Vector3;
  label?: string;
}>();

const sceneStates = inject(SCENE_STATES_KEY)!;

const camera = computed(() => {
  return sceneStates.value!.perspectiveCamera.value;
});

const isOffscreen = ref(false);

const style = reactive({
  transform: "translate(-50%, -50%)",
  left: "0px",
  top: "0px",
  arrowRotation: "0deg",
});

const clonedTargetPos = new Vector3();

watchDebounced(
  [
    sceneStates.value!.currentCam.value.position,
    sceneStates.value!.currentCam.value.rotation,
  ],
  () => {
    const pos = clonedTargetPos.copy(props.targetPos).project(camera.value!);

    // 2. Determine if it is behind the camera
    const isBehind = pos.z > 1;

    // 3. Check boundaries
    const margin = 0.12; // 8% padding from edge
    const limit = 1 - margin;

    // If behind, we flip coordinates to point towards the target correctly
    if (isBehind) {
      pos.x = -pos.x;
      pos.y = -pos.y;
    }

    // Check if actually outside the view frustum
    isOffscreen.value =
      Math.abs(pos.x) > limit || Math.abs(pos.y) > limit || isBehind;

    // 4. Clamping logic
    const clampedX = Math.max(-limit, Math.min(limit, pos.x));
    const clampedY = Math.max(-limit, Math.min(limit, pos.y));

    // 5. Convert to Screen Pixels
    style.left = `${(clampedX * 0.5 + 0.5) * sceneStates.value!.screenSize.width!}px`;
    style.top = `${(-clampedY * 0.5 + 0.5) * sceneStates.value!.screenSize.height!}px`;

    // 6. Calculate rotation for the arrow (only useful when off-screen)
    if (isOffscreen.value) {
      const angle = Math.atan2(pos.y - clampedY, pos.x - clampedX);
      style.arrowRotation = `${Math.PI / 2 - angle}rad`;
    }
  },
  { deep: true, immediate: true, debounce: 10, maxWait: 50 },
);
</script>

<template>
  <div class="beacon-container" :style="style">
    <div :class="['beacon', { 'is-offscreen': isOffscreen }]">
      <div
        v-if="isOffscreen"
        class="arrow absolute origin-center"
        :style="{
          transform: `rotate(${style.arrowRotation}) translateY(-18px)`,
        }"
      >
        <Navigation2 />
      </div>

      <div class="dot">
        <span v-if="!isOffscreen" class="label">{{ label }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.beacon-container {
  position: absolute;
  pointer-events: none;
  z-index: 9;
  will-change: left, top;
}

.beacon {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
}

.dot {
  width: 12px;
  height: 12px;
  background: #00ffcc;
  border: 2px solid white;
  border-radius: 50%;
  box-shadow: 0 0 15px #00ffcc;
}

.label {
  position: absolute;
  top: 20px;
  white-space: nowrap;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.arrow {
  height: 24px;
  font-size: 18px;
  color: #00ffcc;
  margin-bottom: 5px; /* Offset so it points outward */
}

.is-offscreen {
  transform: scale(1.2);
}
</style>
