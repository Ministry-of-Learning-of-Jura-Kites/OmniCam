<script setup lang="ts">
import { Mesh, MeshBasicMaterial } from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

import { ROTATING_TYPE, RotatingUserData } from "./rotating-event-handle";
import { useTresContext } from "@tresjs/core";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import type { MovableObject } from "~/types/movable";
import { useWheelObjGeoCache } from "./use-wheel-obj-geo-cache";
import { CAMERA_UTILS_LAYER } from "~/constants";

const object = defineModel<MovableObject>({ required: true });

const emit = defineEmits<{
  (event: "down" | "up"): void;
  (
    event: "move",
    direction: "x" | "y" | "z",
    directionSign: number,
    delta: number,
  ): void;
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

const sceneStates = inject(SCENE_STATES_KEY)!;

const context = useTresContext();

const { get } = useWheelObjGeoCache();
const geometry = get("wheel");
const material = new MeshBasicMaterial({ color: props.color });
const wheelBase = new Mesh(geometry, material);

wheelBase.layers.set(CAMERA_UTILS_LAYER);

wheelBase.userData = new RotatingUserData(
  props.direction,
  object.value,
  context,
  () => {
    emit("up");
  },
  () => {
    emit("down");
  },
  (direction, angleDir, delta) => {
    emit("move", direction, angleDir, delta);
  },
);

const wheel = wheelBase as unknown as Obj3DWithUserData;

switch (props.direction) {
  case "x":
    wheel.rotateY(Math.PI / 2);
    break;
  case "y":
    wheel.rotateX(Math.PI / 2);
    break;
  case "z":
    break;
  default:
    break;
}

const isActuallyHiding = computed(() => {
  const shouldHide =
    props.isHiding ||
    (object.value.controlling != null &&
      (object.value.controlling.type != ROTATING_TYPE ||
        object.value.controlling.direction != props.direction));

  return shouldHide;
});

function onHidingChange(isHiding: boolean) {
  if (isHiding) {
    sceneStates.value!.draggableObjects.delete(wheel);
  } else {
    sceneStates.value!.draggableObjects.add(wheel);
  }
}

onHidingChange(props.isHiding);

watch(
  () => props.isHiding,
  (newVal) => onHidingChange(newVal),
  { immediate: true },
);

onBeforeUnmount(() => {
  sceneStates.value!.draggableObjects.delete(wheel);
});
</script>

<template>
  <primitive :visible="!isActuallyHiding" :object="wheel" />
</template>
