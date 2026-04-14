<script setup lang="ts">
import { inject, onBeforeUnmount, watchEffect } from "vue";
import { useTresContext } from "@tresjs/core";
import {
  ConeGeometry,
  Group,
  MeshBasicMaterial,
  Mesh,
  SphereGeometry,
  Vector3,
  Matrix4,
} from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";
import { CornerTranslateUserData } from "./corner-translate-handle";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import type { QuadrilateralPoints } from "~/types/trapezoid";

const props = defineProps<{
  faceId: string;
  cornerIndex: number;
  position: Vector3;
  size: number;
  yOffset: number;
}>();

const sceneStates = inject(SCENE_STATES_KEY)!;
const context = useTresContext();

const group = new Group();

// --- Local Assets ---
const arrowLen = props.size;
const arrowRad = props.size * 0.12;

function setNormal(normal: Vector3, points: QuadrilateralPoints) {
  if (points.length < 2) return;

  const localZ = normal.clone().normalize(); // Normal is now Local Z (pointing "out" of the face)

  // Reference X from the first edge
  const p0 = new Vector3(...points[0]);
  const p1 = new Vector3(...points[1]);
  const localX = new Vector3().subVectors(p1, p0).normalize();

  // Calculate Local Y via cross product (X cross Z)
  // This vector lies flat on the plane, perpendicular to the first edge
  const localY = new Vector3().crossVectors(localZ, localX).normalize();

  // Build the rotation matrix: Basis (X, Y, Z)
  const matrix = new Matrix4();
  matrix.makeBasis(localX, localY, localZ);

  group.quaternion.setFromRotationMatrix(matrix);
}

const face = computed(
  () => sceneStates.value?.facesManagement.faces[props.faceId],
);

watch(
  () => face.value!,
  (face) => {
    setNormal(face.normal, face.points);
  },
  {
    immediate: true,
  },
);

const coneGeom = new ConeGeometry(arrowRad, arrowLen, 10);
coneGeom.translate(0, arrowLen / 2, 0); // Offset so pivot is at the base

const sphereGeom = new SphereGeometry(props.size * 0.12, 10, 10);

const matX = new MeshBasicMaterial({
  color: "green",
  transparent: true,
  opacity: 0.9,
  depthTest: false,
});
const matY = new MeshBasicMaterial({
  color: "red",
  transparent: true,
  opacity: 0.9,
  depthTest: false,
});
const matDot = new MeshBasicMaterial({
  color: "white",
  transparent: true,
  opacity: 0.8,
  depthTest: false,
});

const arrows: Obj3DWithUserData[] = [];

// Helper to spawn arrows
function addArrow(axis: "x" | "y", dir: 1 | -1) {
  const mesh = new Mesh(coneGeom, axis === "x" ? matX : matY);

  mesh.rotation.set(0, 0, 0);

  if (axis === "y") {
    // We want the arrow (pointing +Y) to stay on Y
    if (dir === -1) mesh.rotateZ(Math.PI);
  } else {
    // We want the arrow (pointing +Y) to lay down on X
    // Rotate around Z (the normal)
    mesh.rotateZ(dir === 1 ? -Math.PI / 2 : Math.PI / 2);
  }
  mesh.userData = new CornerTranslateUserData(
    axis,
    dir, // Pass direction (1 or -1) to the class
    group,
    props.faceId,
    props.cornerIndex,
    sceneStates.value!,
    context,
    props.yOffset,
  );

  arrows.push(mesh as unknown as Obj3DWithUserData);
  group.add(mesh);
}

// Create the visual structure
const dot = new Mesh(sphereGeom, matDot);
group.add(dot);

addArrow("x", 1);
addArrow("x", -1);
addArrow("y", 1);
addArrow("y", -1);

// --- Reactivity & Dragging ---
watchEffect((onCleanup) => {
  group.position.copy(props.position);

  // Register arrows for the Raycaster/Draggable system
  arrows.forEach((a) => sceneStates.value?.draggableObjects.add(a));

  onCleanup(() => {
    arrows.forEach((a) => sceneStates.value?.draggableObjects.delete(a));
  });
});

// --- Cleanup ---
onBeforeUnmount(() => {
  coneGeom.dispose();
  sphereGeom.dispose();
  matX.dispose();
  matY.dispose();
  matDot.dispose();
});
</script>

<template>
  <primitive :object="group" />
</template>
