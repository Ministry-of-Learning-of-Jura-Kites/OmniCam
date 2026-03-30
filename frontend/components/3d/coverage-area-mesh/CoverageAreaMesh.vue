<!-- OmniCam/frontend/components/3d/coverage-area-mesh/CoverageAreaMesh.vue -->
<script setup lang="ts">
import { computed, watchEffect, onUnmounted } from "vue";
import {
  BufferGeometry,
  Float32BufferAttribute,
  Group,
  LineLoop,
  LineBasicMaterial,
  Mesh,
  MeshBasicMaterial,
  SphereGeometry,
  DoubleSide,
  Matrix4,
  Vector3,
} from "three";
import MovableArrow from "../movable-arrow/MovableArrow.vue";
import TresMesh from "@tresjs/core";
import { AXIS_COLOR } from "~/constants";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";
import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";

type Point3 = [number, number, number];

const model = defineModel<ProcessedCoverageFace>({
  required: false,
})!;

const props = withDefaults(
  defineProps<{
    faceId?: string;
    color?: string;
    previewPoints?: [number, number, number][];
    opacity?: number;
    wireframe?: boolean;
    selected?: boolean;
    showCorners?: boolean;
    yOffset?: number;
  }>(),
  {
    faceId: "",
    color: "#22ff88",
    opacity: undefined,
    wireframe: false,
    selected: false,
    showCorners: true,
    yOffset: 0.01,
    previewPoints: () => [],
  },
);

let isPreview = false;

let pointsRef: Ref<[number, number, number][]>;
if (model.value != null) {
  pointsRef = toRef(model.value, "points");
} else {
  pointsRef = toRef(props, "previewPoints");
  isPreview = true;
}

const center = computed((): Point3 => {
  const sum = pointsRef.value!.reduce(
    (acc, [x, y, z]) => {
      acc[0] += x;
      acc[1] += y;
      acc[2] += z;
      return acc;
    },
    [0, 0, 0] as Point3,
  );

  const count = pointsRef.value!.length;
  return [sum[0] / count, sum[1] / count, sum[2] / count];
});

const fillOpacity = computed(() => {
  if (props.opacity != null) return props.opacity;
  return props.selected ? 0.35 : 0.18;
});
const outlineOpacity = computed(() => (props.selected ? 1 : 0.85));
const cornerRadius = computed(() => (props.selected ? 0.07 : 0.05));

// ---------- THREE objects (stable instances) ----------
const group = new Group();

const meshGeometry = new BufferGeometry();
const outlineGeometry = new BufferGeometry();

const meshMaterial = new MeshBasicMaterial({
  transparent: true,
  opacity: 0.2,
  side: DoubleSide,
  depthWrite: false,
  depthTest: false,
  polygonOffset: true,
  polygonOffsetFactor: -2,
  polygonOffsetUnits: -2,
});

const outlineMaterial = new LineBasicMaterial({
  transparent: true,
  opacity: 1,
});
outlineMaterial.depthTest = false;

const mesh = new Mesh(meshGeometry, meshMaterial);
mesh.renderOrder = 999;
mesh.frustumCulled = false;

const outline = new LineLoop(outlineGeometry, outlineMaterial);
outline.renderOrder = 1000;
outline.frustumCulled = false;

group.add(mesh);
group.add(outline);

const cornerGeom = new SphereGeometry(1, 10, 10);
const cornerMaterial = new MeshBasicMaterial({
  transparent: true,
  opacity: 0.85,
  depthWrite: false,
  depthTest: false,
});

const corners = Array.from({ length: 4 }, () => {
  const m = new Mesh(cornerGeom, cornerMaterial);
  m.renderOrder = 1001;
  m.frustumCulled = false;
  group.add(m);
  return m;
});

const emptyGroup = new Group();
emptyGroup.position.set(center.value[0], center.value[1], center.value[2]);

// Persistent storage for the drag session
let initialPoints: Point3[] | null = null;
function onPosStart() {
  initialPoints = [...pointsRef.value];
}
function onPosDragged(delta: Vector3) {
  if (!pointsRef.value || !pointsRef || !initialPoints) {
    return;
  }
  for (const idx in pointsRef.value) {
    const point = initialPoints[idx];
    pointsRef.value[idx]![0] = point![0] + delta.x;
    pointsRef.value[idx]![1] = point![1] + delta.y;
    pointsRef.value[idx]![2] = point![2] + delta.z;
  }
}

let initialNormal: Point3 | null = null;
let initialCenter: Vector3 | null = null;

function onRotStart() {
  if (!pointsRef.value) return;

  // Deep copy the points so we have a "frozen" reference
  initialPoints = pointsRef.value.map((p) => [...p] as Point3);

  // Store the initial normal
  initialNormal = [...model.value!.normal] as Point3;

  // Calculate the center ONCE at the start of the drag
  const [cx, cy, cz] = center.value;
  initialCenter = new Vector3(cx, cy, cz);
}

function onRotDragged(
  axisName: "x" | "y" | "z",
  directionSign: number,
  totalDelta: number,
) {
  if (
    !pointsRef ||
    !pointsRef.value ||
    !initialPoints ||
    !initialCenter ||
    !initialNormal
  )
    return;

  // 1. Calculate the TOTAL rotation for this gesture
  const rotationMatrix = new Matrix4();
  const axis = new Vector3();
  if (axisName === "x") axis.set(1, 0, 0);
  else if (axisName === "y") axis.set(0, 1, 0);
  else if (axisName === "z") axis.set(0, 0, 1);

  // Apply the full delta from the start of the drag
  rotationMatrix.makeRotationAxis(axis, totalDelta * directionSign);

  // 2. Update the Normal from the initial state
  const nV = new Vector3(...initialNormal);
  nV.applyMatrix4(rotationMatrix).normalize();
  model.value!.normal = [nV.x, nV.y, nV.z];

  // 3. Map initial points to the current pointsRef
  for (let i = 0; i < initialPoints.length; i++) {
    const original = initialPoints[i];
    const v = new Vector3(original![0], original![1], original![2]);

    // Move to initial center -> Rotate -> Move back
    v.sub(initialCenter);
    v.applyMatrix4(rotationMatrix);
    v.add(initialCenter);

    // Write directly to the reactive points array
    pointsRef.value![i]![0] = v.x;
    pointsRef.value![i]![1] = v.y;
    pointsRef.value![i]![2] = v.z;
  }
}

function onRotEnd() {
  // Clean up references
  initialPoints = null;
  initialNormal = null;
  initialCenter = null;
}

function applyUserData() {
  const fid = props.faceId;
  mesh.userData = fid
    ? { kind: "coverage-quad", faceId: fid }
    : { kind: "coverage-quad" };
  outline.userData = fid
    ? { kind: "coverage-outline", faceId: fid }
    : { kind: "coverage-outline" };
  corners.forEach((c, i) => {
    c.userData = fid
      ? { kind: "coverage-corner", faceId: fid, cornerIndex: i }
      : { kind: "coverage-corner", cornerIndex: i };
  });
}
// update geometry positions
function setPositions(geom: BufferGeometry, pts: Point3[]) {
  let attr = geom.getAttribute("position") as
    | Float32BufferAttribute
    | undefined;

  if (!attr || (attr.array as Float32Array).length !== 12) {
    attr = new Float32BufferAttribute(new Float32Array(12), 3);
    geom.setAttribute("position", attr);
  }

  const a = attr.array as Float32Array;

  const p0 = pts[0]!;
  const p1 = pts[1]!;
  const p2 = pts[2]!;
  const p3 = pts[3]!;

  a[0] = p0[0];
  a[1] = p0[1];
  a[2] = p0[2];
  a[3] = p1[0];
  a[4] = p1[1];
  a[5] = p1[2];
  a[6] = p2[0];
  a[7] = p2[1];
  a[8] = p2[2];
  a[9] = p3[0];
  a[10] = p3[1];
  a[11] = p3[2];

  attr.needsUpdate = true;
  geom.computeBoundingSphere();
}

let meshIndexSet = false;

watchEffect(() => {
  const pts = pointsRef.value ?? [];
  const ok = pts.length === 4;

  group.visible = ok;
  if (!ok) return;

  applyUserData();

  const ep: Point3[] = pts.map((p) => [p[0], p[1] + props.yOffset, p[2]]);

  // mesh
  setPositions(meshGeometry, ep);

  if (!meshIndexSet) {
    meshGeometry.setIndex([0, 1, 2, 0, 2, 3]);
    meshIndexSet = true;
  }
  meshGeometry.computeVertexNormals();

  setPositions(outlineGeometry, ep);

  // materials
  meshMaterial.color.set(props.color);
  meshMaterial.opacity = fillOpacity.value;
  meshMaterial.wireframe = !!props.wireframe;

  outlineMaterial.color.set(props.color);
  outlineMaterial.opacity = outlineOpacity.value;

  // corners
  const show = !!props.showCorners;
  const r = cornerRadius.value;
  corners.forEach((c, i) => {
    c.visible = show;
    const p = ep[i]!;
    c.position.set(p[0], p[1], p[2]);
    c.scale.setScalar(r);
  });

  // corner color
  cornerMaterial.color.set(props.selected ? "#ffffff" : props.color);
  cornerMaterial.opacity = props.selected ? 0.95 : 0.8;
});

onUnmounted(() => {
  // remove children
  group.remove(mesh);
  group.remove(outline);
  corners.forEach((c) => group.remove(c));

  // dispose
  meshGeometry.dispose();
  outlineGeometry.dispose();
  cornerGeom.dispose();

  meshMaterial.dispose();
  outlineMaterial.dispose();
  cornerMaterial.dispose();
});
</script>

<template>
  <primitive :object="group" />
  <!-- <TresMesh :position="center" :visible="!isPreview">
    <MovableArrow
      v-for="dir of ['x', 'y', 'z'] as const"
      :key="dir"
      v-model="emptyGroup"
      :direction="dir"
      :color="AXIS_COLOR[dir]"
      @down="onPosStart"
      @move="onPosDragged"
    />
  </TresMesh>
  <TresMesh :position="center" :visible="!isPreview">
    <RotationWheel
      v-for="dir of ['x', 'y', 'z'] as const"
      :key="dir"
      v-model="emptyGroup"
      :direction="dir"
      :color="AXIS_COLOR[dir]"
      @down="onRotStart"
      @move="onRotDragged"
      @up="onRotEnd"
    />
  </TresMesh> -->
</template>
