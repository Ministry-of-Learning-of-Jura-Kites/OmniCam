<script setup lang="ts">
import { Mesh, CylinderGeometry, MeshBasicMaterial, Group } from "three";
import { MOVING_TYPE, MovingUserData } from "./moving-event-handle";
import { MOVING_ARROW_CONFIG } from "~/constants";
import { useTresContext } from "@tresjs/core";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

import type { ICamera } from "~/types/camera";

const cam = defineModel<ICamera>({ required: true });

const emit = defineEmits<{
  (event: "change"): void;
}>();

const props = defineProps({
  direction: {
    type: String as PropType<"x" | "y" | "z">,
    default: "x",
  },
  isHiding: {
    type: Boolean,
    default: false,
  },
  color: {
    type: [String, Number] as PropType<string | number>,
    default: "red",
  },
});

const context = useTresContext();

const sceneStates = inject(SCENE_STATES_KEY)!;

const arrow = new Group();

const cameraUserData = new MovingUserData(
  props.direction,
  cam.value!,
  context!,
  () => {
    emit("change");
  },
);

const material = new MeshBasicMaterial({ color: props.color });

const upHead = new CylinderGeometry(
  0,
  MOVING_ARROW_CONFIG.HEAD_RADIUS,
  MOVING_ARROW_CONFIG.HEAD_LENGTH,
  8,
);
const upHeadMesh = new Mesh(upHead, material);
upHeadMesh.position.y =
  (MOVING_ARROW_CONFIG.CYLINDER_LENGTH + MOVING_ARROW_CONFIG.HEAD_LENGTH) / 2;

const downHead = new CylinderGeometry(
  MOVING_ARROW_CONFIG.HEAD_RADIUS,
  0,
  MOVING_ARROW_CONFIG.HEAD_LENGTH,
  8,
);
const downHeadMesh = new Mesh(downHead, material);
downHeadMesh.position.y =
  -(MOVING_ARROW_CONFIG.CYLINDER_LENGTH + MOVING_ARROW_CONFIG.HEAD_LENGTH) / 2;

const cylinder = new CylinderGeometry(
  MOVING_ARROW_CONFIG.CYLINDER_RADIUS,
  MOVING_ARROW_CONFIG.CYLINDER_RADIUS,
  MOVING_ARROW_CONFIG.CYLINDER_LENGTH,
  8,
);
const cylinderMesh = new Mesh(cylinder, material);

arrow.add(downHeadMesh);
arrow.add(upHeadMesh);
arrow.add(cylinderMesh);

switch (props.direction) {
  case "x":
    arrow.rotateZ(Math.PI / 2);
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
  sceneStates?.draggableObjects.add(obj);
  componentDraggableMeshes.push(obj);
}

function setMeshUndraggable() {
  for (const mesh of componentDraggableMeshes) {
    sceneStates?.draggableObjects.delete(mesh);
  }
}

function setMeshDraggable() {
  for (const mesh of componentDraggableMeshes) {
    sceneStates?.draggableObjects.add(mesh);
  }
}

const isActuallyHiding = computed(() => {
  const shouldHide =
    props.isHiding ||
    (cam.value.controlling != null &&
      (cam.value.controlling.type != MOVING_TYPE ||
        cam.value.controlling.direction != props.direction));

  return shouldHide;
});

function onHidingChange(isHiding: boolean) {
  if (isHiding) {
    setMeshUndraggable();
  } else {
    setMeshDraggable();
  }
}

onHidingChange(props.isHiding);

watch(
  () => props.isHiding,
  (newVal) => onHidingChange(newVal),
  { immediate: true },
);

onBeforeUnmount(() => {
  setMeshUndraggable();
});
</script>

<template>
  <primitive :visible="!isActuallyHiding" :object="arrow" />
</template>
