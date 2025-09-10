<script setup lang="ts">
import * as THREE from "three";
import { SCENE_STATES_KEY } from "../scene-states-provider/create-scene-states";
import { RotatingUserData } from "./rotating-event-handle";
import { useTresContext } from "@tresjs/core";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { ROTATING_TORUS_CONFIG } from "~/constants";
import type { ICamera } from "~/types/camera";

const cam = defineModel<ICamera>({ required: true });

const props = defineProps({
  type: {
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

wheelBase.userData = new RotatingUserData(props.type, cam.value, context);

const wheel = wheelBase as unknown as Obj3DWithUserData;

switch (props.type) {
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
</script>

<template>
  <primitive :object="wheel" />
</template>
