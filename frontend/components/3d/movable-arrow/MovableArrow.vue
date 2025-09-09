<script setup lang="ts">
import * as THREE from "three";
import { MovingUserData } from "./moving-event-handle";
import { MOVING_ARROW_CONFIG } from "~/constants";
import { useTresContext } from "@tresjs/core";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { SCENE_STATES_KEY } from "~/components/3d/scene-states-provider/create-scene-states";

const props = defineProps({
  direction: {
    type: String as PropType<"x" | "y" | "z">,
    default: "x",
  },
  cameraMesh: {
    type: Object as PropType<THREE.Mesh | null>,
    default: null,
  },
  color: {
    type: [String, Number] as PropType<string | number>,
    default: "red",
  },
  draggableObjects: {
    type: Set<THREE.Object3D>,
    default: new Set(),
  },
});

const context = useTresContext();

const sceneStates = inject(SCENE_STATES_KEY)!;

const arrow = new THREE.Group();

const cameraUserData = new MovingUserData(
  props.direction,
  props.cameraMesh!,
  context!,
);

const material = new THREE.MeshBasicMaterial({ color: props.color });

const upHead = new THREE.CylinderGeometry(
  0,
  MOVING_ARROW_CONFIG.HEAD_RADIUS,
  MOVING_ARROW_CONFIG.HEAD_LENGTH,
  8,
);
const upHeadMesh = new THREE.Mesh(upHead, material);
upHeadMesh.position.y =
  (MOVING_ARROW_CONFIG.CYLINDER_LENGTH + MOVING_ARROW_CONFIG.HEAD_LENGTH) / 2;

const downHead = new THREE.CylinderGeometry(
  MOVING_ARROW_CONFIG.HEAD_RADIUS,
  0,
  MOVING_ARROW_CONFIG.HEAD_LENGTH,
  8,
);
const downHeadMesh = new THREE.Mesh(downHead, material);
downHeadMesh.position.y =
  -(MOVING_ARROW_CONFIG.CYLINDER_LENGTH + MOVING_ARROW_CONFIG.HEAD_LENGTH) / 2;

const cylinder = new THREE.CylinderGeometry(
  MOVING_ARROW_CONFIG.CYLINDER_RADIUS,
  MOVING_ARROW_CONFIG.CYLINDER_RADIUS,
  MOVING_ARROW_CONFIG.CYLINDER_LENGTH,
  8,
);
const cylinderMesh = new THREE.Mesh(cylinder, material);

arrow.add(downHeadMesh);
arrow.add(upHeadMesh);
arrow.add(cylinderMesh);

switch (props.direction) {
  case "x":
    arrow.rotateZ(Math.PI / 2);
    // arrow.updateMatrixWorld(true);
    break;
  case "y":
    break;
  case "z":
    arrow.rotateX(Math.PI / 2);
    // arrow.updateMatrixWorld(true);
    break;
  default:
    break;
}

const componentDraggableMeshes: Obj3DWithUserData[] = [];

for (const mesh of arrow.children) {
  const obj = mesh as Obj3DWithUserData;
  obj.userData = cameraUserData;
  sceneStates.draggableObjects.add(obj);
  componentDraggableMeshes.push(obj);
}

onUnmounted(() => {
  for (const mesh of componentDraggableMeshes) {
    sceneStates.draggableObjects.delete(mesh);
  }
});
</script>

<template>
  <primitive :object="arrow" />
</template>
