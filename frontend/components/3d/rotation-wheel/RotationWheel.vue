<script setup lang="ts">
import { Mesh, MeshBasicMaterial, TorusGeometry } from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

import { ROTATING_TYPE, RotatingUserData } from "./rotating-event-handle";
import { useTresContext } from "@tresjs/core";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { ROTATING_TORUS_CONFIG } from "~/constants";
import type { ICamera } from "~/types/camera";

const cam = defineModel<ICamera>({ required: true });

const props = defineProps({
  direction: {
    type: String as PropType<"x" | "y" | "z">,
    default: "x",
  },
  cameraMesh: {
    type: Object as PropType<Mesh | null>,
    default: null,
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

const sceneStates = inject(SCENE_STATES_KEY);

const context = useTresContext();

const geometry = new TorusGeometry(
  ROTATING_TORUS_CONFIG.RADIUS,
  ROTATING_TORUS_CONFIG.TUBE_RADIUS,
  16,
  100,
);
const material = new MeshBasicMaterial({ color: props.color });
const wheelBase = new Mesh(geometry, material);

wheelBase.userData = new RotatingUserData(props.direction, cam.value, context);

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
    (cam.value.controlling != null &&
      (cam.value.controlling.type != ROTATING_TYPE ||
        cam.value.controlling.direction != props.direction));

  return shouldHide;
});

function onHidingChange(isHiding: boolean) {
  if (isHiding) {
    sceneStates?.draggableObjects.delete(wheel);
  } else {
    sceneStates?.draggableObjects.add(wheel);
  }
}

onHidingChange(props.isHiding);

watch(
  () => props.isHiding,
  (newVal) => onHidingChange(newVal),
  { immediate: true },
);

onBeforeUnmount(() => {
  sceneStates?.draggableObjects.delete(wheel);
});
</script>

<template>
  <primitive :visible="!isActuallyHiding" :object="wheel" />
</template>
