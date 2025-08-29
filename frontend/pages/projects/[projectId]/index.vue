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
  <div class="container py-10 mx-auto">
    <DataTable :columns="generateKey" :data="dataArray" />
  </div>
  <Form
    v-model:open="isFormDialogOpen"
    :fields="fields"
    :model="models"
    :isUpdate="isCreateOrUpdate"
    @submit="handleFormSubmit"
  />
</template>
