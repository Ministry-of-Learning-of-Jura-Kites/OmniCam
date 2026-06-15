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

const sceneStates = inject(SCENE_STATES_KEY)!;

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

const state = shallowRef<GLTF | null>(null);
const data = shallowRef<ArrayBuffer | null>(null);

onMounted(async () => {
  if (!BufferGeometry.prototype.computeBoundsTree) {
    BufferGeometry.prototype.computeBoundsTree = computeBoundsTree;
    BufferGeometry.prototype.disposeBoundsTree = disposeBoundsTree;
    Mesh.prototype.raycast = acceleratedRaycast;
  }

  // fetch model data without top-level await — no Suspense conflict
  try {
    const resp = await fetch(props.path ?? "", {
      method: "GET",
      credentials: "include",
      cache: "no-cache",
    });

    if (!resp.ok) {
      emit("err", `Failed to load 3D model (HTTP ${resp.status}).`);
      return;
    }

    data.value = await resp.arrayBuffer();
  } catch (error: unknown) {
    console.error(error);
  }

  if (!data.value) return;

  const blob = new Blob([data.value], { type: "model/gltf-binary" });
  const blobUrl = URL.createObjectURL(blob);

  const gltf = useGLTF(blobUrl);

  const stopInner = watch(
    () => gltf.state.value,
    async (s) => {
      if (s != undefined) {
        state.value = s;

        s.scene.traverse((child) => {
          child.layers.enable(MINIMAP_LAYER);
          if ((child as Mesh).isMesh) {
            const mesh = child as Mesh;
            mesh.geometry.computeBoundsTree();
            mesh.raycast = acceleratedRaycast;
          }
        });

        sceneStates.value!.modelRef.value = s;

        URL.revokeObjectURL(blobUrl);
        stopInner();
      }
    },
  );
});

onUnmounted(() => {
  if (state?.value?.scene) {
    state.value.scene.traverse((child: Object3D) => {
      if (isMesh(child)) {
        const mesh = child as Mesh;
        mesh.geometry?.dispose();

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
  <primitive v-if="state?.scene" ref="mesh" :object="state.scene" />

  <!-- Block Placeholder while loading -->
  <TresMesh v-if="state?.scene == null" :position="props.position ?? [0, 0, 0]">
    <TresBoxGeometry />
    <TresMeshStandardMaterial
      :color="'#4a90e2'"
      :metalness="0.3"
      :roughness="0.4"
    />
  </TresMesh>
</template>
