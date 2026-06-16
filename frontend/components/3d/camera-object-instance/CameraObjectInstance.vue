<script setup lang="ts">
import { InstancedMesh, MeshBasicMaterial, Object3D } from "three";
import type { ICamera } from "~/types/camera";
import { useCamObjGeoCache } from "./use-cam-obj-geo-cache";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { SCENE_STATES_KEY } from "~/constants/state-keys";
import { CAMERA_UTILS_LAYER } from "~/constants";
import CameraInstanceTracker from "../camera-instance-tracker/CameraInstanceTracker.vue";

const props = defineProps<{
  cameras: Record<string, ICamera>;
  color?: string;
}>();

const cameraIds = computed(() => Object.keys(props.cameras));

const MAX_CAMERAS = 1000;
const sceneStates = inject(SCENE_STATES_KEY)!;
const { get } = useCamObjGeoCache();

const bodyInstancedMesh = new InstancedMesh(
  get("body"),
  new MeshBasicMaterial({ color: props.color ?? "white" }),
  MAX_CAMERAS,
);
bodyInstancedMesh.layers.set(CAMERA_UTILS_LAYER);
bodyInstancedMesh.frustumCulled = false;
bodyInstancedMesh.userData = { type: "camera-instanced", cameraIds };

const lensInstancedMesh = new InstancedMesh(
  get("lens"),
  new MeshBasicMaterial({ color: "black" }),
  MAX_CAMERAS,
);
lensInstancedMesh.layers.set(CAMERA_UTILS_LAYER);
lensInstancedMesh.frustumCulled = false;
lensInstancedMesh.userData = { type: "camera-instanced", cameraIds };

const dummy = new Object3D();
const camIndexMap = new Map<string, number>();

function updateInstance(cam: ICamera, i: number) {
  dummy.position.copy(cam.position);
  dummy.quaternion.setFromEuler(cam.rotation);
  dummy.updateMatrix();
  bodyInstancedMesh.setMatrixAt(i, dummy.matrix);
  lensInstancedMesh.setMatrixAt(i, dummy.matrix);
}

function flushMatrices() {
  bodyInstancedMesh.instanceMatrix.needsUpdate = true;
  lensInstancedMesh.instanceMatrix.needsUpdate = true;
}

function updateCameraMatrix(id: string) {
  const i = camIndexMap.get(id);
  if (i === undefined) return;
  updateInstance(props.cameras[id]!, i);
  flushMatrices();
}

defineExpose({ cameraIds, updateCameraMatrix });

watch(
  () => Object.keys(props.cameras),
  (ids, oldIds) => {
    const oldSet = new Set(oldIds ?? []);

    const hasStructuralChange =
      ids.length !== (oldIds?.length ?? 0) || ids.some((id) => !oldSet.has(id));

    if (!hasStructuralChange) return;

    camIndexMap.clear();
    bodyInstancedMesh.count = ids.length;
    lensInstancedMesh.count = ids.length;

    ids.forEach((id, i) => {
      camIndexMap.set(id, i);
      if (!oldSet.has(id)) {
        updateInstance(props.cameras[id]!, i);
      }
    });

    flushMatrices();
  },
  { immediate: true },
);

onMounted(() => {
  sceneStates.value!.clickableObjects.add(
    bodyInstancedMesh as unknown as Obj3DWithUserData,
  );
  sceneStates.value!.clickableObjects.add(
    lensInstancedMesh as unknown as Obj3DWithUserData,
  );
});

onUnmounted(() => {
  sceneStates.value!.clickableObjects.delete(
    bodyInstancedMesh as unknown as Obj3DWithUserData,
  );
  sceneStates.value!.clickableObjects.delete(
    lensInstancedMesh as unknown as Obj3DWithUserData,
  );
});
</script>

<template>
  <primitive :object="bodyInstancedMesh" />
  <primitive :object="lensInstancedMesh" />
  <CameraInstanceTracker
    v-for="(cam, id) in cameras"
    :key="id"
    :cam-id="String(id)"
    :cam="cam"
    :on-update="updateCameraMatrix"
  />
</template>
