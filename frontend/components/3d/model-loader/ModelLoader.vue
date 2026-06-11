<script setup lang="ts">
import { useGLTF } from "@tresjs/cientos";
import { BufferGeometry, Mesh, type Material, type Object3D } from "three";
import type { GLTF } from "three-stdlib";
import { SCENE_STATES_KEY } from "~/constants/state-keys";
import {
  computeBoundsTree,
  disposeBoundsTree,
  acceleratedRaycast,
} from "three-mesh-bvh";
import { MINIMAP_LAYER } from "~/constants";
// import { init as initRecast, Crowd } from "recast-navigation";
// import { threeToSoloNavMesh, NavMeshHelper } from "@recast-navigation/three";
// import { Crowd } from "recast-navigation";

const sceneStates = inject(SCENE_STATES_KEY)!;
// const mesh = ref<Mesh>();

const props = withDefaults(
  defineProps<{
    path: string;
    position?: [number, number, number];
    modelScale?: number;
  }>(),
  {
    modelScale: 1,
    position: () => [0, 0, 0],
  },
);
const emit = defineEmits<{
  (e: "err", message: string): void;
}>();

function isMesh(object: Object3D): object is Mesh {
  return (object as Mesh).isMesh === true;
}

function disposeMaterial(mat: Material) {
  for (const key in mat) {
    const value = mat[key as keyof Material];
    if (
      value &&
      typeof value === "object" &&
      "dispose" in value &&
      typeof value.dispose === "function"
    ) {
      value.dispose();
    }
  }
  mat.dispose();
}

console.log("path", props.path);
const state = shallowRef<GLTF | null>(null);
const { data, error, status } = await useFetch<ArrayBuffer>(props.path ?? "", {
  method: "GET",
  credentials: "include",
  responseType: "arrayBuffer",
  cache: "no-cache",
});

watch(
  error,
  (err) => {
    if (err) {
      emit("err", `Failed to load 3D model (HTTP ${status.value}).`);
    }
  },
  { immediate: true },
);

onMounted(() => {
  if (!BufferGeometry.prototype.computeBoundsTree) {
    BufferGeometry.prototype.computeBoundsTree = computeBoundsTree;
    BufferGeometry.prototype.disposeBoundsTree = disposeBoundsTree;
    Mesh.prototype.raycast = acceleratedRaycast;
  }

  const stopWatch = watch(
    data,
    (newData) => {
      if (newData) {
        const blob = new Blob([newData], { type: "model/gltf-binary" });
        const blobUrl = URL.createObjectURL(blob);

        const gltf = useGLTF(blobUrl);
        const stopInner = watch(
          () => gltf.state.value,
          async (s) => {
            if (s != undefined) {
              // s.scene.traverse(applyFisheye);
              state.value = s;
              // const walkableMeshes: Mesh[] = [];
              s.scene.traverse((child) => {
                child.layers.enable(MINIMAP_LAYER);
                if ((child as Mesh).isMesh) {
                  const mesh = child as Mesh;
                  mesh.geometry.computeBoundsTree();
                  // Optional: Helps with raycasting through complex hierarchies
                  mesh.raycast = acceleratedRaycast;
                }
              });

              sceneStates.value!.modelRef.value = s;

              // s.scene.traverse((child) => {
              //   if ((child as Mesh).isMesh) {
              //     walkableMeshes.push(child as Mesh);
              //   }
              // });

              // try {
              //   await initRecast(); // WASM init — safe to call multiple times

              //   const { navMesh, success } = threeToSoloNavMesh(
              //     walkableMeshes,
              //     {
              //       cs: 0.1, // finer voxel size
              //       ch: 0.02, // finer height resolution

              //       walkableSlopeAngle: 35,

              //       walkableHeight: 1.5,
              //       walkableClimb: 0.5,

              //       walkableRadius: 0.1,

              //       maxEdgeLen: 32,
              //       maxSimplificationError: 2,

              //       minRegionArea: 0,
              //       maxVertsPerPoly: 6,

              //       detailSampleDist: 1,
              //       detailSampleMaxError: 0.05,
              //     },
              //   );
              //   if (!success || !navMesh) {
              //     console.error("[ModelLoader] NavMesh build failed");
              //   } else {
              //     console.log("[ModelLoader] NavMesh ready ✓");

              //     sceneStates.value!.navMesh.value = navMesh;
              //     const crowd = new Crowd(navMesh, {
              //       maxAgents: 20,
              //       maxAgentRadius: 0.3,
              //     });

              //     sceneStates.value!.crowd.value = crowd;

              //     const helper = new NavMeshHelper(navMesh);
              //     s.scene.add(helper);
              //   }
              // } catch (e) {
              //   console.error("[ModelLoader] NavMesh init error:", e);
              // }

              URL.revokeObjectURL(blobUrl!);

              stopInner();
              stopWatch();
            }
          },
        );
      }
    },
    { immediate: true },
  );
});

onUnmounted(() => {
  // sceneStates.value?.crowd?.value?.destroy();
  // sceneStates.value?.navMesh?.value?.destroy();
  if (state?.value?.scene) {
    state.value.scene.traverse((child: Object3D) => {
      if (isMesh(child)) {
        const mesh = child as Mesh;
        // Dispose Geometries
        mesh.geometry?.dispose();

        // Dispose Materials
        if (Array.isArray(mesh.material)) {
          mesh.material.forEach((mat: Material) => disposeMaterial(mat));
        } else {
          disposeMaterial(mesh.material);
        }
      }
    });

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    state.value.scene = null as any;
  }
  state.value = null;
});
</script>

<template>
  <!-- <primitive v-if="state?.scene" ref="mesh" :object="state.scene" /> -->
  <primitive v-if="state?.scene" ref="mesh" :object="state.scene" />

  <!-- Block Placeholder  -->
  <TresMesh v-if="state?.scene == null" :position="props.position ?? [0, 0, 0]">
    <TresBoxGeometry />
    <TresMeshStandardMaterial
      :color="'#4a90e2'"
      :metalness="0.3"
      :roughness="0.4"
    />
  </TresMesh>
</template>
