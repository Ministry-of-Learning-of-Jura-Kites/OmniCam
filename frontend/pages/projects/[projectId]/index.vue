<script setup lang="ts">
// import { keyof } from "zod";
import { generateColumnsFromKeys } from "~/components/dataTable/column";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue";
import FormDialog from "~/components/dialog/FormDialog.vue";

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

export interface ModelReturnRequest {
  data: Model;
}

export interface ModelCreateRequest {
  name: string;
  description: string;
  file: File;
}

export interface ModelUpdateRequest {
  name: string;
  description: string;
}

export type ModelForm = {
  name: string;
  description: string;
  file: File | null;
};
export type ModelWithoutId = Omit<Model, "id">;

const config = useRuntimeConfig();
const route = useRoute();

const models = ref<Record<string, ModelWithoutId>>({});
const currentEditId = ref<string | null>(null);
const modelForm = reactive<ModelForm>({
  name: "",
  description: "",
  file: null,
});

// write table model (what u want to show etc.)
const modelKeys: (keyof Model)[] = [
  "name",
  "description",
  "version",
  "createdAt",
  "updatedAt",
];

const tableTitles = {
  name: "Name",
  description: "Description",
  version: "Version",
  createdAt: "Created At",
  updatedAt: "Updated At",
};

// generate key to use here

const generateKey = generateColumnsFromKeys<Model>(modelKeys, tableTitles, {
  onEdit: (row) => handleEditRow(row),
  onDelete: (row) => handleDeleteRow(row),
});

// Dialog handler for create delete update
const isEditFormDialogOpen = ref<boolean>(false);
const isCreateFormDialogOpen = ref<boolean>(false);
const confirmDialog = ref<boolean>(false);
const confirmMessage = ref<string>("");
const successDialog = ref<boolean>(false);
// const isLoading = ref(false);

// dialog form config
// type is input type (text , number , textarea , file etc)
const editfields = {
  name: "text",
  description: "textarea",
  // file: "file",
} as const;

const createFields = {
  name: "text",
  description: "textarea",
  file: "file",
} as const;

const formTitles = {
  name: "Name",
  description: "Description",
  file: "Model file",
};

async function fetchModel() {
  const projectId = route.params.projectId as string;
  try {
    const response = await $fetch<ModelGetRequest>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models`,
      {
        method: "GET",
      },
    );

    models.value = response.data.reduce<Record<string, ModelWithoutId>>(
      (acc, model) => {
        const { id, ...rest } = model;
        acc[id] = rest;
        return acc;
      },
      {},
    );

    console.log(models.value, "c");
  } catch (err) {
    console.error("fetchModel error", err);
  }
}

async function createModel() {
  const formData = new FormData();
  formData.append("name", modelForm.name);
  formData.append("description", modelForm.description);
  if (modelForm.file) {
    formData.append("file", modelForm.file);
  }

  const projectId = route.params.projectId as string;
  const response = await $fetch<ModelReturnRequest>(
    `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models`,
    {
      method: "POST",
      body: formData,
    },
  );

  models.value[response.data.id] = {
    name: response.data.name,
    description: response.data.description,
    updatedAt: response.data.updatedAt,
    createdAt: response.data.createdAt,
    version: response.data.version,
  } as ModelWithoutId;
  successDialog.value = true;
}

async function updateModel(modelId: string) {
  const projectId = route.params.projectId as string;
  console.log(modelId);
  console.log(modelForm);
  try {
    const response = await $fetch<ModelReturnRequest>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models/${modelId}`,
      {
        method: "PUT",
        body: {
          name: modelForm.name,
          description: modelForm.description,
        },
      },
    );
    console.log(models.value);
    console.log(response);
    models.value = {
      ...models.value,
      [modelId]: {
        name: response.data.name,
        description: response.data.description,
        version: response.data.version,
        createdAt: response.data.createdAt,
        updatedAt: response.data.updatedAt,
        projectId: response.data.projectId,
      },
    };
    console.log("new Model", models.value[modelId]);
    successDialog.value = true;
    console.log("Updated row:", modelId);
  } catch (err) {
    console.error("Update failed", err);
  }
}

async function deleteRow(id: string) {
  try {
    const projectId = route.params.projectId as string;

    await $fetch(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models/${id}`,
      {
        method: "DELETE",
      },
    );

    models.value = Object.fromEntries(
      Object.entries(models.value).filter(([key]) => key !== id),
    );
    console.log("Deleted row:", id);
    successDialog.value = true;
  } catch (err) {
    console.error("Delete failed", err);
  }
}
function handleCreate() {
  modelForm.name = "";
  modelForm.description = "";
  modelForm.file = null;
  isCreateFormDialogOpen.value = true;
}

function handleEditRow(row: Model) {
  currentEditId.value = row.id;
  confirmMessage.value = `Do you want to update this ${row.name}`;
  isEditFormDialogOpen.value = true;
  modelForm.description = row.description;
  modelForm.name = row.name;
  modelForm.file = null;
}

function handleDeleteRow(row: Model) {
  currentEditId.value = row.id;
  confirmDialog.value = true;
  confirmMessage.value = `Do you want to delete this ${row.name}`;
}

// Form submit handler
function handleEditFormSubmit() {
  console.log("a");
  isEditFormDialogOpen.value = false;
  confirmDialog.value = true;
  console.log(confirmDialog.value);
}

function handleConfirmSubmit() {
  if (!currentEditId.value) {
    return;
  }

  if (confirmMessage.value.includes("delete")) {
    deleteRow(currentEditId.value);
  } else {
    console.log("is in");
    updateModel(currentEditId.value);
  }

  confirmDialog.value = false;
}

function handleCreateFormSubmit() {
  isCreateFormDialogOpen.value = false;
  createModel();
}

onMounted(() => {
  fetchModel();
});

// data computed when change (record -> array)
const dataArray = computed(() =>
  Object.entries(models.value).map(([id, model]) => ({ id, ...model })),
);

watch(
  modelForm,
  (modelForm) => {
    console.log(modelForm);
  },
  {
    once: false,
  },
);
</script>

<template>
  <div>
    <Button type="button" @click="handleCreate" />
    <div class="container py-10 mx-auto">
      <DataTable :columns="generateKey" :data="dataArray" />
    </div>
    <FormDialog
      v-model:open="isEditFormDialogOpen"
      v-model:model="modelForm"
      mode="update"
      :fields="editfields"
      :titles="formTitles"
      @submit="handleEditFormSubmit"
    />

    <FormDialog
      v-model:open="isCreateFormDialogOpen"
      v-model:model="modelForm"
      mode="create"
      :fields="createFields"
      :titles="formTitles"
      @submit="handleCreateFormSubmit"
    />
    <ConfirmDialog
      v-model:open="confirmDialog"
      :message="confirmMessage"
      @submit="handleConfirmSubmit"
    />
  </div>
</template>
