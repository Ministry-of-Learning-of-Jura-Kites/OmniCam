<script setup lang="ts">
import { CSG } from "three-csg-ts";
import type { BufferGeometry } from "three";
import {
  Mesh,
  Quaternion,
  MeshStandardMaterial,
  MeshBasicMaterial,
  Group,
  DoubleSide,
  Box3,
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

const tempMaterial = new MeshStandardMaterial();
const tempMeshA = new Mesh(undefined, tempMaterial);
const tempMeshB = new Mesh(undefined, tempMaterial);
const boxA = new Box3();
const boxB = new Box3();
const overlayGroup = new Group();
const workingQuaternion = new Quaternion();

function syncMeshToCamera(
  mesh: Mesh,
  geometry: BufferGeometry,
  camState: ICamera,
) {
  mesh.geometry = geometry;
  mesh.position.copy(camState.position);
  workingQuaternion.setFromEuler(camState.rotation);
  mesh.quaternion.copy(workingQuaternion);
  mesh.updateMatrixWorld(true);
}

function updateOrAddMesh(newGeometry: BufferGeometry, index: number) {
  if (index < overlayGroup.children.length) {
    const existingMesh = overlayGroup.children[index] as Mesh;
    existingMesh.geometry.dispose(); // Critical disposal
    existingMesh.geometry = newGeometry;
    existingMesh.visible = true;
  } else {
    const newMesh = new Mesh(newGeometry, intersectMaterial);
    overlayGroup.add(newMesh);
  }
}

function updateIntersections() {
  // Hide current pool
  overlayGroup.children.forEach((child) => (child.visible = false));

  const cameraPairs = visibleFrustum.value;
  let poolIndex = 0;

  for (let i = 0; i < cameraPairs.length; i++) {
    for (let j = i + 1; j < cameraPairs.length; j++) {
      const [idA, camA] = cameraPairs[i]!;
      const [idB, camB] = cameraPairs[j]!;

      const geomA = getFrustumGeometry(idA)?.mesh;
      const geomB = getFrustumGeometry(idB)?.mesh;

      if (!geomA || !geomB) continue;

      // Sync worker meshes
      syncMeshToCamera(tempMeshA, geomA, camA);
      syncMeshToCamera(tempMeshB, geomB, camB);

      boxA.setFromObject(tempMeshA);
      boxB.setFromObject(tempMeshB);

      // Early exit if bounds don't touch
      if (!boxA.intersectsBox(boxB)) continue;

      // Bake transformation (Matches your old working code)
      const bakedGeomA = geomA.clone().applyMatrix4(tempMeshA.matrixWorld);
      const bakedGeomB = geomB.clone().applyMatrix4(tempMeshB.matrixWorld);

      const bakedMeshA = new Mesh(bakedGeomA);
      const bakedMeshB = new Mesh(bakedGeomB);

      const intersectMesh = CSG.intersect(bakedMeshA, bakedMeshB);

      bakedGeomA.dispose();
      bakedGeomB.dispose();

      const geo = intersectMesh.geometry;

      if (geo.attributes.position && geo.attributes.position.count > 0) {
        updateOrAddMesh(geo, poolIndex);
        poolIndex++;
      } else {
        geo.dispose(); // Cleanup failed intersection
      }
    }
  }
}

onMounted(() => {
  watch(
    () => visibleFrustum.value,
    () => {
      nextTick(updateIntersections);
    },
    {
      deep: true,
      immediate: true,
    },
  );
});

onUnmounted(() => {
  overlayGroup.children.forEach((child) => {
    if (child instanceof Mesh) {
      child.geometry.dispose();
    }
  });
  overlayGroup.clear();
  tempMaterial.dispose();
});
</script>

<template>
  <primitive :object="overlayGroup" />
</template>
