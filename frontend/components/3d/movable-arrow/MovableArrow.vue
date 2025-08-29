<script setup lang="ts">
import * as THREE from "three";
import { CameraUserData } from "../camera-object/camera-user-data";
import { ARROW_CONFIG } from "~/constants";
import { useTresContext } from "@tresjs/core";
import type { ObjWithUserData } from "../scene-3d/obj-event-handler";
import { SCENE_STATES_KEY } from "../scene-3d/use-scene-state";

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

const cameraUserData = new CameraUserData(
  props.direction,
  props.cameraMesh!,
  context!,
);

const material = new THREE.MeshBasicMaterial({ color: props.color });

const upHead = new THREE.CylinderGeometry(
  0,
  ARROW_CONFIG.HEAD_RADIUS,
  ARROW_CONFIG.HEAD_LENGTH,
  8,
);
const upHeadMesh = new THREE.Mesh(upHead, material);
upHeadMesh.position.y = ARROW_CONFIG.CYLINDER_LENGTH / 2;

const downHead = new THREE.CylinderGeometry(
  ARROW_CONFIG.HEAD_RADIUS,
  0,
  ARROW_CONFIG.HEAD_LENGTH,
  8,
);
const downHeadMesh = new THREE.Mesh(downHead, material);
downHeadMesh.position.y = -ARROW_CONFIG.CYLINDER_LENGTH / 2;

const cylinder = new THREE.CylinderGeometry(
  ARROW_CONFIG.CYLINDER_RADIUS,
  ARROW_CONFIG.CYLINDER_RADIUS,
  ARROW_CONFIG.CYLINDER_LENGTH,
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

const componentDraggableMeshes: ObjWithUserData[] = [];

for (const mesh of arrow.children) {
  const obj = mesh as ObjWithUserData;
  obj.userData = cameraUserData;
  sceneStates.draggableObjects.add(obj);
  console.log(sceneStates.draggableObjects);
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
