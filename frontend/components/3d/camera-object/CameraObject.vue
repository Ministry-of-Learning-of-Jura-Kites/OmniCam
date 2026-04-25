<script setup lang="ts">
import { SCENE_STATES_KEY } from "@/constants/state-keys";

import MovableArrow from "../movable-arrow/MovableArrow.vue";
import TresMesh from "@tresjs/core";
import RotationWheel from "../rotation-wheel/RotationWheel.vue";
import CameraFrustum from "../camera-frustum/CameraFrustum.vue";
import type { Color, Group } from "three";
import { Quaternion } from "three";
import { safeGetAspectRatio } from "~/utils/aspect-ratio";
import type { ICamera } from "~/types/camera";
import { useCamObjGeoCache } from "./use-cam-obj-geo-cache";

const props = withDefaults(
  defineProps<{
    name?: string;
    camId: string;
    workspace: string | null;
    color?: string | number | Color | [r: number, g: number, b: number];
    instance?: ICamera;
  }>(),
  {
    name: "Untitled",
    color: "white",
    instance: undefined,
  },
);

const { get: getGeo } = useCamObjGeoCache();

const cameraBodyGeo = getGeo("body");
const cameraLensGeo = getGeo("lens");

const sceneStates = inject(SCENE_STATES_KEY)!;

const group = ref<Group>();

let cam: Ref<ICamera>;
if (props.instance == undefined) {
  cam = toRef(sceneStates.value!.cameras, props.camId);
} else {
  cam = ref(props.instance);
}

const camQuat = computed(() => {
  const quaternion = new Quaternion().setFromEuler(cam!.value.rotation);
  return quaternion;
});

watch(group, (group) => {
  group!.layers.enable(1);
});
</script>

<template>
  <TresGroup
    ref="group"
    :visible="sceneStates!.currentCamId.value !== props.camId"
    :position="[cam!.position.x, cam!.position.y, cam!.position.z]"
  >
    <TresObject3D :quaternion="camQuat">
      <TresMesh :geometry="cameraBodyGeo">
        <TresMeshBasicMaterial :color="props.color" />
      </TresMesh>
      <TresMesh :geometry="cameraLensGeo">
        <TresMeshBasicMaterial :color="'black'" />
      </TresMesh>
      <CameraFrustum
        :id="camId"
        :fov="cam!.fov"
        :aspect="safeGetAspectRatio(cam.widthRes, cam.heightRes)"
        :length="cam!.frustumLength"
        :color="cam!.frustumColor"
        :is-hiding="
          cam!.isHidingFrustum || camId == sceneStates!.currentCamId.value
        "
      />
    </TresObject3D>
    <template v-if="cam != null">
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isLockingPosition ||
          sceneStates!.currentCamId.value == props.camId ||
          props.workspace != 'me'
        "
        :controlling="cam.controlling"
        direction="x"
        color="green"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isLockingPosition ||
          sceneStates!.currentCamId.value == props.camId ||
          props.workspace != 'me'
        "
        :controlling="cam.controlling"
        direction="y"
        color="red"
      />
      <MovableArrow
        v-model="cam"
        :is-hiding="
          cam.isLockingPosition ||
          sceneStates!.currentCamId.value == props.camId ||
          props.workspace != 'me'
        "
        :controlling="cam.controlling"
        direction="z"
        color="blue"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isLockingRotation ||
          sceneStates!.currentCamId.value == props.camId ||
          props.workspace != 'me'
        "
        direction="x"
        color="green"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isLockingRotation ||
          sceneStates!.currentCamId.value == props.camId ||
          props.workspace != 'me'
        "
        direction="y"
        color="red"
      />
      <RotationWheel
        v-model="cam"
        :is-hiding="
          cam.isLockingRotation ||
          sceneStates!.currentCamId.value == props.camId ||
          props.workspace != 'me'
        "
        direction="z"
        color="blue"
      />
    </template>
  </TresGroup>
</template>
