<script setup lang="ts">
import * as THREE from "three";
import { SCENE_STATES_KEY } from "../scene-states-provider/create-scene-states";
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
    type: Object as PropType<THREE.Mesh | null>,
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

const geometry = new THREE.TorusGeometry(
  ROTATING_TORUS_CONFIG.RADIUS,
  ROTATING_TORUS_CONFIG.TUBE_RADIUS,
  16,
  100,
);
const material = new THREE.MeshBasicMaterial({ color: props.color });
const wheelBase = new THREE.Mesh(geometry, material);

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

sceneStates?.draggableObjects.add(wheel);

const isActuallyHiding = computed(() => {
  const shouldHide =
    props.isHiding ||
    (cam.value.controlling != null &&
      (cam.value.controlling.type != ROTATING_TYPE ||
        cam.value.controlling.direction != props.direction));

  return shouldHide;
});

watch(
  () => props.isHiding,
  (isHiding) => {
    if (isHiding) {
      sceneStates?.draggableObjects.delete(wheel);
    } else {
      sceneStates?.draggableObjects.add(wheel);
    }
  },
);

onBeforeUnmount(() => {
  sceneStates?.draggableObjects.delete(wheel);
});
</script>

<template>
  <primitive :visible="!isActuallyHiding" :object="wheel" />
</template>
