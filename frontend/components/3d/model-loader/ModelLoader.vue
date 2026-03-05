<script setup lang="ts">
import { useGLTF } from "@tresjs/cientos";
import type { Material, Mesh, Object3D } from "three";
import type { GLTF } from "three-stdlib";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

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

onMounted(() => {
  const stopWatch = watch(
    data,
    (newData) => {
      if (newData) {
        const blob = new Blob([newData], { type: "model/gltf-binary" });
        const blobUrl = URL.createObjectURL(blob);

        const gltf = useGLTF(blobUrl);
        const stopInner = watch(
          () => gltf.state.value,
          (s) => {
            if (s != undefined) {
              // s.scene.traverse(applyFisheye);
              state.value = s;

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

  watch(error, () => {
    if (error.value != undefined) {
      throw new Error(`HTTP ${status.value} ${error.value}`);
    }
  });
});

onUnmounted(() => {
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

const sceneStates = inject(SCENE_STATES_KEY);

const finalScale = computed(() => {
  // return props.modelScale * sceneStates!.calibration.scale;
  return props.modelScale;
});
</script>

<template>
  <TresGroup
    :position="[
      props.position?.[0] ?? 0,
      (props.position?.[1] ?? 0) + sceneStates!.calibration.heightOffset,
      props.position?.[2] ?? 0,
    ]"
    :scale="[finalScale, finalScale, finalScale]"
  >
    <primitive v-if="state?.scene" :object="state.scene" />
  </TresGroup>

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
