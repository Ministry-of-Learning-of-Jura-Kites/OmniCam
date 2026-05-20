<script setup lang="ts">
import { Vector3, type Object3D, type PerspectiveCamera } from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

const props = defineProps<{
  perspectiveCamera: PerspectiveCamera | null;
  getSurfaceHit: (
    event: PointerEvent,
  ) => { point: Vector3; normal: Vector3 } | null;
}>();

const sceneStates = inject(SCENE_STATES_KEY)!;

const MEASUREMENT_PALETTE = [
  "#f97316", // orange
  "#a855f7", // purple
  "#06b6d4", // cyan
  "#f59e0b", // amber
  "#ec4899", // pink
  "#84cc16", // lime
  "#14b8a6", // teal
  "#fb923c", // light orange
];

const lineColors = reactive<Record<string, string>>({});

function getNextMeasurementColor(): string {
  const used = Object.keys(lineColors).length;
  return MEASUREMENT_PALETTE[used % MEASUREMENT_PALETTE.length]!;
}

watch(
  () => sceneStates.value!.measurement.lines.length,
  () => {
    for (const line of sceneStates.value!.measurement.lines) {
      if (!lineColors[line.id]) {
        lineColors[line.id] = getNextMeasurementColor();
      }
    }
    for (const id of Object.keys(lineColors)) {
      if (!sceneStates.value!.measurement.lines.find((l) => l.id === id)) {
        delete lineColors[id];
      }
    }
  },
);

const measurementDotRegistry = new Map<
  Object3D,
  { lineId: string; pointIndex: 0 | 1 }
>();

const draggingMeasurementPoint = ref<{
  lineId: string;
  pointIndex: 0 | 1;
} | null>(null);

const draftDotMeshRef = ref<unknown>(null);
const draggingDraftPoint = ref(false);

function registerDot(el: unknown, lineId: string, pointIndex: 0 | 1) {
  if (el) measurementDotRegistry.set(el as Object3D, { lineId, pointIndex });
}

function unregisterDot(el: unknown) {
  if (el) measurementDotRegistry.delete(el as Object3D);
}

// ── Labels ───────────────────────────────────────────────────────────
const labelStyles = ref<Record<string, Record<string, string>>>({});

function worldToScreen(position: Vector3) {
  const camera = props.perspectiveCamera;
  if (
    !camera ||
    !sceneStates.value?.screenSize.width ||
    !sceneStates.value?.screenSize.height
  ) {
    return null;
  }

  const projected = position.clone().project(camera);

  if (projected.z > 1) return null;

  return {
    x: (projected.x + 1) * 0.5 * sceneStates.value.screenSize.width,
    y: (-projected.y + 1) * 0.5 * sceneStates.value.screenSize.height,
  };
}

function getMeasurementLabelStyle(line: {
  start: Vector3;
  end: Vector3;
}): Record<string, string> {
  const worldPos = line.start.clone().add(new Vector3(0, 0.08, 0));
  const screen = worldToScreen(worldPos);

  if (!screen) return { display: "none" };

  return {
    display: "block",
    left: `${screen.x}px`,
    top: `${screen.y}px`,
    transform: "translate(-50%, -100%)",
  };
}

function updateLabelStyles() {
  if (!sceneStates.value?.measurement?.lines?.length) return;

  const updated: typeof labelStyles.value = {};
  for (const line of sceneStates.value.measurement.lines) {
    updated[line.id] = getMeasurementLabelStyle(line);
  }
  labelStyles.value = updated;
}

let labelRafId: number | null = null;

function labelLoop() {
  updateLabelStyles();
  labelRafId = requestAnimationFrame(labelLoop);
}

onMounted(() => {
  labelRafId = requestAnimationFrame(labelLoop);
});

onUnmounted(() => {
  if (labelRafId !== null) {
    cancelAnimationFrame(labelRafId);
    labelRafId = null;
  }
});

function createLinePoints(start: Vector3, end: Vector3) {
  return new Float32Array([start.x, start.y, start.z, end.x, end.y, end.z]);
}

function onPointerEvent(
  event: PointerEvent,
  //   eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
  raycaster: { intersectObjects: Function },
): boolean {
  if (event.type === "pointerup") {
    if (draggingDraftPoint.value || draggingMeasurementPoint.value) {
      draggingDraftPoint.value = false;
      draggingMeasurementPoint.value = null;
      return true;
    }
    return false;
  }

  if (event.type === "pointermove") {
    if (draggingDraftPoint.value) {
      const hit = props.getSurfaceHit(event);
      if (hit) sceneStates.value!.measurement.draftStartPoint = hit.point;
      return true;
    }
    if (draggingMeasurementPoint.value) {
      const hit = props.getSurfaceHit(event);
      if (hit) {
        sceneStates.value!.measurement.updateEndpoint(
          draggingMeasurementPoint.value.lineId,
          draggingMeasurementPoint.value.pointIndex,
          hit.point,
        );
      }
      return true;
    }
    return false;
  }

  if (event.type === "pointerdown") {
    if (
      draftDotMeshRef.value &&
      sceneStates.value!.measurement.draftStartPoint
    ) {
      const draftHits = raycaster.intersectObjects(
        [draftDotMeshRef.value as Object3D],
        false,
      );
      if (draftHits.length > 0) {
        draggingDraftPoint.value = true;
        return true;
      }
    }

    const dotMeshes = [...measurementDotRegistry.keys()];
    if (dotMeshes.length > 0) {
      const dotHits = raycaster.intersectObjects(dotMeshes, false);
      if (dotHits.length > 0) {
        const meta = measurementDotRegistry.get(dotHits[0]!.object);
        if (meta) {
          draggingMeasurementPoint.value = meta;
          return true;
        }
      }
    }
  }

  return false;
}

defineExpose({ onPointerEvent });
</script>

<template>
  <template
    v-for="line in sceneStates!.measurement.lines"
    :key="`line-${line.id}-${line.start.x.toFixed(4)}-${line.start.z.toFixed(4)}-${line.end.x.toFixed(4)}-${line.end.z.toFixed(4)}`"
  >
    <TresLine>
      <TresBufferGeometry>
        <TresBufferAttribute
          attach="attributes-position"
          :array="createLinePoints(line.start, line.end)"
          :count="2"
          :item-size="3"
        />
      </TresBufferGeometry>
      <TresLineBasicMaterial
        :color="lineColors[line.id] ?? '#00ff88'"
        :linewidth="2"
      />
    </TresLine>
  </template>

  <!-- Dots -->
  <template
    v-for="line in sceneStates!.measurement.lines"
    :key="`points-${line.id}`"
  >
    <TresMesh
      :ref="
        (el: unknown) => (el ? registerDot(el, line.id, 0) : unregisterDot(el))
      "
      :position="line.start"
    >
      <TresSphereGeometry :args="[0.05, 16, 16]" />
      <TresMeshBasicMaterial :color="lineColors[line.id] ?? '#00ff88'" />
    </TresMesh>

    <TresMesh
      :ref="
        (el: unknown) => (el ? registerDot(el, line.id, 1) : unregisterDot(el))
      "
      :position="line.end"
    >
      <TresSphereGeometry :args="[0.05, 16, 16]" />
      <TresMeshBasicMaterial :color="lineColors[line.id] ?? '#00ff88'" />
    </TresMesh>
  </template>

  <!-- Draft start dot -->
  <TresMesh
    v-if="sceneStates!.measurement.draftStartPoint"
    :ref="
      (el: unknown) => {
        draftDotMeshRef = el;
      }
    "
    :position="sceneStates!.measurement.draftStartPoint"
  >
    <TresSphereGeometry :args="[0.05, 16, 16]" />
    <TresMeshBasicMaterial
      :color="
        MEASUREMENT_PALETTE[
          Object.keys(lineColors).length % MEASUREMENT_PALETTE.length
        ]
      "
      :opacity="0.7"
      :transparent="true"
    />
  </TresMesh>
</template>
