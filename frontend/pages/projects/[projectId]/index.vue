<script lang="ts">
export interface Model {
  id: string;
  projectId: string;
  name: string;
  description: string;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface ModelCreateRequest {
  name: string;
  description: string;
}

export interface ModelUpdateRequest {
  name: string;
  description: string;
}
</script>
<template>
  <header>
    list all models including create model / update model (not camera btw)
    pathing to /models/:modelId
  </header>
</template>
