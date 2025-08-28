<script setup lang="ts">
export interface Model {
  id: string;
  projectId: string;
  name: string;
  description: string;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface ModelGetRequest {
  data: Model[];
}

export interface ModelCreateRequest {
  name: string;
  description: string;
}

export interface ModelUpdateRequest {
  name: string;
  description: string;
}
const config = useRuntimeConfig();
const route = useRoute();

const models = ref<Model[]>();

async function fetchModel() {
  const projectId = route.params.projectId as string;
  console.log(projectId, "project id");

  const response = await $fetch<ModelGetRequest>(
    `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models`,
    {
      method: "GET",
    },
  );

  models.value = response.data;
}

onMounted(() => {
  fetchModel();
});
</script>
<template>
  <div v-for:="model in models">
    <h1>{{ model.name }}</h1>
  </div>
  <!-- <header>
    list all models including create model / update model (not camera btw)
    pathing to /models/:modelId
  </header> -->
</template>
