<script setup lang="ts">
import { watchEffect } from "vue";
import { CSG } from "three-csg-ts";
import {
  Mesh,
  Quaternion,
  MeshStandardMaterial,
  MeshBasicMaterial,
  Group,
} from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";
import { useFrustumGeometries } from "~/composables/useFrustumGeometries";
import type { ICamera } from "~/types/camera";

const sceneStates = inject(SCENE_STATES_KEY)!;

const visibleFrustum = computed(() =>
  Object.entries(sceneStates.cameras).filter(
    ([_, cam]: [string, ICamera]) => !cam.isHidingFrustum,
  ),
);

const { getFrustumGeometry } = useFrustumGeometries();

const overlayGroup = new Group();

watchEffect(() => {
  overlayGroup.children.forEach((obj) => {
    if (obj instanceof Mesh) {
      obj.geometry.dispose();
      if (Array.isArray(obj.material)) {
        obj.material.forEach((m) => m.dispose());
      } else {
        obj.material.dispose();
      }
    }
  });
  overlayGroup.clear();

  const cameraIds = visibleFrustum.value.map(([id]) => id);

  for (let i = 0; i < cameraIds.length; i++) {
    for (let j = i + 1; j < cameraIds.length; j++) {
      const geometryA = getFrustumGeometry(cameraIds[i]!)?.mesh;
      const geometryB = getFrustumGeometry(cameraIds[j]!)?.mesh;

      if (!geometryA || !geometryB) continue;

      const meshA = new Mesh(geometryA, new MeshStandardMaterial());
      const meshB = new Mesh(geometryB, new MeshStandardMaterial());

      const camA = visibleFrustum.value.find(
        ([key]) => key === cameraIds[i],
      )?.[1];
      const camB = visibleFrustum.value.find(
        ([key]) => key === cameraIds[j],
      )?.[1];

      if (camA) {
        meshA.position.copy(camA.position);
      }
      meshA.quaternion.copy(new Quaternion().setFromEuler(camA!.rotation));
      meshA.updateMatrixWorld(true);

      if (camB) {
        meshB.position.copy(camB.position);
      }
      meshB.quaternion.copy(new Quaternion().setFromEuler(camB!.rotation));
      meshB.updateMatrixWorld(true);

      const intersectMesh = CSG.intersect(meshA, meshB);

      if (intersectMesh.geometry.attributes.position!.count > 0) {
        intersectMesh.material = new MeshBasicMaterial({
          color: 0xff0000,
          transparent: true,
          opacity: 0.5,
        });
        overlayGroup.add(intersectMesh);
      }
    }
  }
});
</script>

<template>
  <primitive :object="overlayGroup" />
</template>
