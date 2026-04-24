<script setup lang="ts">
import type { Vector3 } from "three";
import { Mesh, MeshBasicMaterial, Group } from "three";
import { MOVING_TYPE, MovingUserData } from "./moving-event-handle";
import { useTresContext } from "@tresjs/core";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

import type { MovableObject } from "~/types/movable";
import { useArrowObjGeoCache } from "./use-arrows-obj-geo-cache";

const object = defineModel<MovableObject>({ required: true });

const emit = defineEmits<{
  (event: "down"): void;
  (event: "move", delta: Vector3): void;
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
  object.value,
  context!,
  () => {
    emit("down");
  },
  (delta) => {
    emit("move", delta);
  },
);

const material = new MeshBasicMaterial({ color: props.color });

const { get } = useArrowObjGeoCache();
const geo = get("arrow");
const arrowMesh = new Mesh(geo, material);

arrow.add(arrowMesh);

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
  sceneStates.value!.draggableObjects.add(obj);
  componentDraggableMeshes.push(obj);
}

function setMeshUndraggable() {
  for (const mesh of componentDraggableMeshes) {
    sceneStates.value!.draggableObjects.delete(mesh);
  }
}

function setMeshDraggable() {
  for (const mesh of componentDraggableMeshes) {
    sceneStates.value!.draggableObjects.add(mesh);
  }
}

const isActuallyHiding = computed(() => {
  const shouldHide =
    props.isHiding ||
    (object.value.controlling != null &&
      (object.value.controlling.type != MOVING_TYPE ||
        object.value.controlling.direction != props.direction));

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
