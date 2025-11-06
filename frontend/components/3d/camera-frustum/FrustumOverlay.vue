<script setup lang="ts">
import { watchEffect } from "vue";
import { CSG } from "three-csg-ts";
import * as THREE from "three";
import { SCENE_STATES_KEY } from "~/components/3d/scene-states-provider/create-scene-states";
import { useFrustumGeometries } from "~/composables/useFrustumGeometries";
import { type ICamera } from "~/types/camera";

const sceneStates = inject(SCENE_STATES_KEY)!;

const visibleFrustum = computed(() =>
  Object.entries(sceneStates.cameras).filter(
    ([_, cam]: [string, ICamera]) => !cam.isHidingFrustum,
  ),
);

const { getFrustumGeometry } = useFrustumGeometries();

const overlayGroup = new THREE.Group();

watchEffect(() => {
  overlayGroup.clear();

  const cameraIds = visibleFrustum.value.map(([id]) => id);

  for (let i = 0; i < cameraIds.length; i++) {
    for (let j = i + 1; j < cameraIds.length; j++) {
      const geometryA = getFrustumGeometry(cameraIds[i]!)?.mesh;
      const geometryB = getFrustumGeometry(cameraIds[j]!)?.mesh;

      if (!geometryA || !geometryB) continue;

      const meshA = new THREE.Mesh(geometryA, new THREE.MeshStandardMaterial());
      const meshB = new THREE.Mesh(geometryB, new THREE.MeshStandardMaterial());

      const intersectMesh = CSG.intersect(meshA, meshB);

      if (intersectMesh.geometry.attributes.position!.count > 0) {
        intersectMesh.material = new THREE.MeshBasicMaterial({
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
