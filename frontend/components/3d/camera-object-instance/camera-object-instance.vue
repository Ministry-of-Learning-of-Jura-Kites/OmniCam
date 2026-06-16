<script setup lang="ts">
import { computed, watch } from "vue";
import { InstancedMesh, MeshBasicMaterial, Object3D } from "three";

import type { ICamera } from "~/types/camera";
import { useCamObjGeoCache } from "./use-cam-obj-geo-cache";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { SCENE_STATES_KEY } from "~/constants/state-keys";
import { CAMERA_UTILS_LAYER } from "~/constants";

const props = defineProps<{
  cameras: Record<string, ICamera>;
  color?: string;
}>();

const cameraIds = computed(() => Object.keys(props.cameras));

defineExpose({
  cameraIds,
});

const MAX_CAMERAS = 1000;
const sceneStates = inject(SCENE_STATES_KEY)!;
const { get } = useCamObjGeoCache();
const cameraBodyGeo = get("body");
const cameraLensGeo = get("lens");

const bodyMaterial = new MeshBasicMaterial({
  color: props.color ?? "white",
});

const lensMaterial = new MeshBasicMaterial({
  color: "black",
});

const bodyInstancedMesh = new InstancedMesh(
  cameraBodyGeo,
  bodyMaterial,
  MAX_CAMERAS,
);
bodyInstancedMesh.layers.set(CAMERA_UTILS_LAYER);
bodyInstancedMesh.frustumCulled = false;

const lensInstancedMesh = new InstancedMesh(
  cameraLensGeo,
  lensMaterial,
  MAX_CAMERAS,
);
lensInstancedMesh.layers.set(CAMERA_UTILS_LAYER);
lensInstancedMesh.frustumCulled = false;

const dummy = new Object3D();

const cameraList = computed(() => Object.values(props.cameras));

bodyInstancedMesh.userData = {
  type: "camera-instanced",
  cameraIds,
};

lensInstancedMesh.userData = {
  type: "camera-instanced",
  cameraIds,
};

watch(
  cameraList,
  (cams) => {
    bodyInstancedMesh.count = cams.length;
    lensInstancedMesh.count = cams.length;

    for (let i = 0; i < cams.length; i++) {
      const cam = cams[i];

      dummy.position.copy(cam!.position);
      dummy.quaternion.setFromEuler(cam!.rotation);
      dummy.updateMatrix();

      bodyInstancedMesh.setMatrixAt(i, dummy.matrix);
      lensInstancedMesh.setMatrixAt(i, dummy.matrix);
    }

    bodyInstancedMesh.instanceMatrix.needsUpdate = true;
    lensInstancedMesh.instanceMatrix.needsUpdate = true;
  },
  { immediate: true },
);

const cameraTransforms = computed(() =>
  Object.values(props.cameras).map((cam) => ({
    px: cam.position.x,
    py: cam.position.y,
    pz: cam.position.z,
    rx: cam.rotation.x,
    ry: cam.rotation.y,
    rz: cam.rotation.z,
  })),
);

watch(
  cameraTransforms,
  () => {
    const cams = Object.values(props.cameras);
    bodyInstancedMesh.count = cams.length;
    lensInstancedMesh.count = cams.length;

    for (let i = 0; i < cams.length; i++) {
      const cam = cams[i]!;
      dummy.position.copy(cam.position);
      dummy.quaternion.setFromEuler(cam.rotation);
      dummy.updateMatrix();

      bodyInstancedMesh.setMatrixAt(i, dummy.matrix);
      lensInstancedMesh.setMatrixAt(i, dummy.matrix);
    }

    bodyInstancedMesh.instanceMatrix.needsUpdate = true;
    lensInstancedMesh.instanceMatrix.needsUpdate = true;
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
</template>
