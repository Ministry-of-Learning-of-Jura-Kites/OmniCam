<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from "vue";
import { Vector3, Euler, Matrix4 } from "three";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const sceneStates = inject(SCENE_STATES_KEY);
const gizmoCanvas = ref<HTMLCanvasElement | null>(null);

const SIZE = 128;
const CENTER = SIZE / 2;
const ARM_LENGTH = 44;
const DOT_R = 11;
const NEG_DOT_R = 11;
const EPS = 0.0001;

type AxisKey = "x" | "y" | "z";

interface ActiveState {
  key: AxisKey;
  isNeg: boolean;
  label: string;
}

const activeView = ref<ActiveState | null>(null);

const AXES = [
  {
    key: "x" as const,
    vec: new Vector3(1, 0, 0),
    color: "#E85555",
    dim: "#6B2222",
  },
  {
    key: "y" as const,
    vec: new Vector3(0, 1, 0),
    color: "#5ABF5A",
    dim: "#1F5C1F",
  },
  {
    key: "z" as const,
    vec: new Vector3(0, 0, 1),
    color: "#5588E8",
    dim: "#1F3470",
  },
];

const VIEW_LABELS: Record<AxisKey, { pos: string; neg: string }> = {
  x: { pos: "RIGHT (+X)", neg: "LEFT (-X)" },
  y: { pos: "TOP (+Y)", neg: "BOTTOM (-Y)" },
  z: { pos: "FRONT (+Z)", neg: "BACK (-Z)" },
};

const PRESETS: Record<
  AxisKey,
  {
    pos: { x: number; y: number; z: number };
    neg: { x: number; y: number; z: number };
  }
> = {
  x: {
    pos: { x: 0, y: Math.PI / 2, z: 0 },
    neg: { x: 0, y: -Math.PI / 2, z: 0 },
  },
  y: {
    pos: { x: -Math.PI / 2, y: 0, z: 0 },
    neg: { x: Math.PI / 2, y: 0, z: 0 },
  },
  z: {
    pos: { x: 0, y: 0, z: 0 },
    neg: { x: 0, y: Math.PI, z: 0 },
  },
};

let rafId: number | null = null;

interface ProjectedEnd {
  key: AxisKey;
  isNeg: boolean;
  sx: number;
  sy: number;
  depth: number;
  color: string;
  dim: string;
}

function project(rotX: number, rotY: number, rotZ: number): ProjectedEnd[] {
  const inv = new Matrix4()
    .makeRotationFromEuler(new Euler(rotX, rotY, rotZ, "YXZ"))
    .invert();

  const result: ProjectedEnd[] = [];
  for (const axis of AXES) {
    const vPos = axis.vec.clone().applyMatrix4(inv);
    result.push({
      key: axis.key,
      isNeg: false,
      sx: CENTER + vPos.x * ARM_LENGTH,
      sy: CENTER - vPos.y * ARM_LENGTH,
      depth: vPos.z,
      color: axis.color,
      dim: axis.dim,
    });
    const vNeg = axis.vec.clone().negate().applyMatrix4(inv);
    result.push({
      key: axis.key,
      isNeg: true,
      sx: CENTER + vNeg.x * ARM_LENGTH,
      sy: CENTER - vNeg.y * ARM_LENGTH,
      depth: vNeg.z,
      color: axis.color,
      dim: axis.dim,
    });
  }
  return result;
}

function draw() {
  const canvas = gizmoCanvas.value;
  const ctx = canvas?.getContext("2d");
  if (!canvas || !ctx) {
    rafId = requestAnimationFrame(draw);
    return;
  }

  const rot = sceneStates?.value?.currentCam.value.rotation;
  const ends = project(rot?.x ?? 0, rot?.y ?? 0, rot?.z ?? 0);

  ctx.clearRect(0, 0, SIZE, SIZE);

  // Background circle
  ctx.beginPath();
  ctx.arc(CENTER, CENTER, CENTER - 1, 0, Math.PI * 2);
  ctx.fillStyle = "rgba(20, 18, 42, 0.55)";
  ctx.fill();

  // Active axis ring on background circle edge
  if (activeView.value) {
    const activeAxis = AXES.find((a) => a.key === activeView.value!.key)!;
    ctx.beginPath();
    ctx.arc(CENTER, CENTER, CENTER - 1, 0, Math.PI * 2);
    ctx.strokeStyle = activeAxis.color;
    ctx.lineWidth = 2;
    ctx.globalAlpha = 0.6;
    ctx.stroke();
    ctx.globalAlpha = 1;
  }

  const sorted = [...ends].sort((a, b) => a.depth - b.depth);

  // Arms (neg → pos, drawn under dots)
  for (const end of sorted) {
    if (end.isNeg) continue;
    const negEnd = ends.find((e) => e.key === end.key && e.isNeg)!;
    ctx.beginPath();
    ctx.moveTo(negEnd.sx, negEnd.sy);
    ctx.lineTo(end.sx, end.sy);
    ctx.strokeStyle = end.depth > negEnd.depth ? end.color : end.dim;
    ctx.lineWidth = 2.5;
    ctx.globalAlpha = 0.8;
    ctx.stroke();
    ctx.globalAlpha = 1;
  }

  // Dots + labels
  for (const end of sorted) {
    const isFront = end.depth >= -EPS; // treat near-zero depth as front-facing to avoid flicker
    const r = end.isNeg ? NEG_DOT_R : DOT_R;
    const fillColor = isFront ? end.color : end.dim;
    const alpha = isFront ? 1 : 0.5;

    const isActive =
      activeView.value?.key === end.key &&
      activeView.value?.isNeg === end.isNeg;

    ctx.globalAlpha = alpha;

    // Dot fill
    ctx.beginPath();
    ctx.arc(end.sx, end.sy, r, 0, Math.PI * 2);
    ctx.fillStyle = isActive ? "white" : fillColor;
    ctx.fill();

    // Active dot: colored ring + axis-color inner circle so text still reads
    if (isActive) {
      // outer white ring
      ctx.strokeStyle = fillColor;
      ctx.lineWidth = 2.5;
      ctx.stroke();
      // inner colored fill so label contrasts
      ctx.beginPath();
      ctx.arc(end.sx, end.sy, r - 3, 0, Math.PI * 2);
      ctx.fillStyle = fillColor;
      ctx.fill();
    }

    // Label
    if (!end.isNeg || isFront) {
      ctx.fillStyle = "white";
      ctx.font = "bold 11px sans-serif";
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      const label = end.isNeg
        ? `-${end.key.toUpperCase()}`
        : end.key.toUpperCase();
      ctx.fillText(label, end.sx, end.sy);
    }

    ctx.globalAlpha = 1;
  }

  // View label at bottom of gizmo circle
  if (activeView.value) {
    const axisColor = AXES.find((a) => a.key === activeView.value!.key)!.color;
    ctx.font = "bold 8px sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillStyle = axisColor;
    ctx.globalAlpha = 0.95;
    ctx.fillText(activeView.value.label, CENTER, SIZE - 9);
    ctx.globalAlpha = 1;
  }

  rafId = requestAnimationFrame(draw);
}

function onGizmoClick(e: MouseEvent) {
  const canvas = gizmoCanvas.value;
  if (!canvas) return;

  const rect = canvas.getBoundingClientRect();
  const mx = (e.clientX - rect.left) * (SIZE / rect.width);
  const my = (e.clientY - rect.top) * (SIZE / rect.height);

  const rot = sceneStates?.value?.currentCam.value.rotation;
  const ends = project(rot?.x ?? 0, rot?.y ?? 0, rot?.z ?? 0);
  const frontFirst = [...ends].sort((a, b) => b.depth - a.depth);

  for (const end of frontFirst) {
    const r = end.isNeg ? NEG_DOT_R + 4 : DOT_R + 4;
    const dist = Math.hypot(mx - end.sx, my - end.sy);
    if (dist <= r) {
      const cam = sceneStates?.value?.currentCam.value;
      const preset = end.isNeg ? PRESETS[end.key].neg : PRESETS[end.key].pos;
      if (!cam) return;
      cam.rotation.x = preset.x;
      cam.rotation.y = preset.y;
      cam.rotation.z = preset.z;
      // ← set active so gizmo highlights it
      activeView.value = {
        key: end.key,
        isNeg: end.isNeg,
        label: VIEW_LABELS[end.key][end.isNeg ? "neg" : "pos"],
      };
      return;
    }
  }
}

onMounted(() => {
  rafId = requestAnimationFrame(draw);
});
onBeforeUnmount(() => {
  if (rafId !== null) cancelAnimationFrame(rafId);
});
</script>

<template>
  <canvas
    ref="gizmoCanvas"
    :width="SIZE"
    :height="SIZE"
    class="axis-gizmo"
    title="Click axis to snap view"
    @click="onGizmoClick"
  />
</template>

<style scoped>
.axis-gizmo {
  position: absolute;
  bottom: 16px;
  left: 16px;
  width: 128px;
  height: 128px;
  z-index: 20;
  cursor: pointer;
  border-radius: 50%;
  transition: filter 0.15s;
}
.axis-gizmo:hover {
  filter: brightness(1.2);
}
</style>
