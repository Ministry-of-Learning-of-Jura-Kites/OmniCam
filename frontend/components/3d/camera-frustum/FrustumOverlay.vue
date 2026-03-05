<script setup lang="ts">
import { watchEffect } from "vue";
import { CSG } from "three-csg-ts";
import {
  Mesh,
  Quaternion,
  MeshStandardMaterial,
  MeshBasicMaterial,
  Group,
  DoubleSide,
} from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";
import { useFrustumGeometries } from "~/composables/useFrustumGeometries";
import type { ICamera } from "~/types/camera";

const sceneStates = inject(SCENE_STATES_KEY)!;
const intersectMaterial = new MeshBasicMaterial({
  color: 0xff0000,
  transparent: true,
  opacity: 0.5,
  side: DoubleSide,
  polygonOffset: true,
  polygonOffsetFactor: -1,
  polygonOffsetUnits: -1,
  depthTest: true,
});

const visibleFrustum = computed(() =>
  Object.entries(sceneStates.cameras).filter(
    ([_, cam]: [string, ICamera]) => !cam.isHidingFrustum,
  ),
);

const { getFrustumGeometry } = useFrustumGeometries();

const overlayGroup = new Group();

watchEffect(() => {
  overlayGroup.children.forEach((obj) => {
    obj.visible = false;
  });

  const cameraIds = visibleFrustum.value.map(([id]) => id);
  let poolIndex = 0;

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

      meshA.geometry.computeBoundingBox();
      meshB.geometry.computeBoundingBox();

      if (meshA.geometry.boundingBox && meshB.geometry.boundingBox) {
        const boxA = meshA.geometry.boundingBox
          .clone()
          .applyMatrix4(meshA.matrixWorld);
        const boxB = meshB.geometry.boundingBox
          .clone()
          .applyMatrix4(meshB.matrixWorld);

        if (!boxA.intersectsBox(boxB)) {
          continue;
        }

        const tempGeomA = meshA.geometry.clone();
        const tempGeomB = meshB.geometry.clone();

        tempGeomA.applyMatrix4(meshA.matrixWorld);
        tempGeomB.applyMatrix4(meshB.matrixWorld);

        const tempMeshA = new Mesh(tempGeomA, meshA.material);
        const tempMeshB = new Mesh(tempGeomB, meshB.material);

        const intersectMesh = CSG.intersect(tempMeshA, tempMeshB);

        tempGeomA.dispose();
        tempGeomB.dispose();

        if (intersectMesh.geometry.attributes.position!.count > 0) {
          if (poolIndex < overlayGroup.children.length) {
            const displayMesh = overlayGroup.children[poolIndex] as Mesh;
            displayMesh.geometry.dispose();
            displayMesh.geometry = intersectMesh.geometry;
            displayMesh.visible = true;
            displayMesh.renderOrder = 0;
          } else {
            intersectMesh.material = intersectMaterial;
            overlayGroup.add(intersectMesh);
          }
          poolIndex++;
        }
      }
    }
  }
});
</script>

<template>
  <primitive :object="overlayGroup" />
</template>
