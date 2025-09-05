<script setup lang="ts">
import * as THREE from "three";
import { SCENE_STATES_KEY } from "../scene-states-provider/create-scene-states";
import { RotatingUserData } from "./rotating-event-handle";
import { useTresContext } from "@tresjs/core";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { ROTATING_TORUS_CONFIG } from "~/constants";

const props = defineProps({
  color: {
    type: String,
    required: true,
  },
  type: {
    type: String,
    required: true,
  },
  objRef: {
    type: THREE.Mesh,
    required: true,
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

wheelBase.userData = new RotatingUserData(props.type, props.objRef, context);

const wheel = wheelBase as unknown as Obj3DWithUserData;

sceneStates?.draggableObjects.add(wheel);
</script>

<template>
  <primitive :object="wheel" />
</template>
