<script setup lang="ts">
import { Navigation2 } from "lucide-vue-next";
import { Vector3 } from "three";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const props = defineProps<{
  targetPos: Vector3;
  label?: string;
  show: boolean;
}>();

const sceneStates = inject(SCENE_STATES_KEY)!;

const camera = computed(() => sceneStates.value!.perspectiveCamera.value);

const isOffscreen = ref(false);

const style = reactive({
  left: "0px",
  top: "0px",
  arrowRotation: "0rad",
});

const clonedTargetPos = new Vector3();

let rafId: number | null = null;

const tick = () => {
  rafId = null;

  if (!props.show) return;

  const pos = clonedTargetPos.copy(props.targetPos).project(camera.value!);

  const isBehind = pos.z > 1;

  const margin = 0.12;
  const limit = 1 - margin;

  if (isBehind) {
    pos.x = -pos.x;
    pos.y = -pos.y;
  }

  isOffscreen.value =
    Math.abs(pos.x) > limit || Math.abs(pos.y) > limit || isBehind;

  const clampedX = Math.max(-limit, Math.min(limit, pos.x));
  const clampedY = Math.max(-limit, Math.min(limit, pos.y));

  style.left = `${(clampedX * 0.5 + 0.5) * sceneStates.value!.screenSize.width!}px`;
  style.top = `${(-clampedY * 0.5 + 0.5) * sceneStates.value!.screenSize.height!}px`;

  if (isOffscreen.value) {
    const angle = Math.atan2(pos.y - clampedY, pos.x - clampedX);
    style.arrowRotation = `${Math.PI / 2 - angle}rad`;
  }
};

const scheduleUpdate = () => {
  if (rafId !== null) return;
  rafId = requestAnimationFrame(tick);
};

watch(
  () => [
    sceneStates.value!.currentCam.value.position,
    sceneStates.value!.currentCam.value.rotation,
    props.targetPos,
    sceneStates.value!.screenSize,
    props.show,
  ],
  scheduleUpdate,
  { deep: true, immediate: true },
);
</script>

<template>
  <div
    v-if="props.show"
    class="beacon-container"
    :style="{ left: style.left, top: style.top }"
  >
    <!-- off-screen: pill with arrow + label -->
    <template v-if="isOffscreen">
      <div class="pill">
        <Navigation2
          class="pill-arrow"
          :style="{ transform: `rotate(${style.arrowRotation})` }"
        />
        <span v-if="label" class="pill-label">{{ label }}</span>
      </div>
    </template>

    <!-- on-screen: dot with label below -->
    <template v-else>
      <div class="on-screen">
        <div class="dot" />
        <span v-if="label" class="label">{{ label }}</span>
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
  will-change: left, top;
}

/* ── off-screen pill ───────────────────────────────────────── */
.pill {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px 6px 8px;
  background: rgba(0, 0, 0, 0.75);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 999px;
  backdrop-filter: blur(6px);
}

.pill-arrow {
  width: 16px;
  height: 16px;
  color: #00ffcc;
  flex-shrink: 0;
  transition: transform 0.12s ease;
}

.pill-label {
  color: white;
  font-size: 12px;
  white-space: nowrap;
}

/* ── on-screen dot + label ─────────────────────────────────── */
.on-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
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
  white-space: nowrap;
  background: rgba(0, 0, 0, 0.75);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
</style>
