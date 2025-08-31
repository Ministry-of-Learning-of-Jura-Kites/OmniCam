<script setup lang="ts">
// import { keyof } from "zod";
import { generateColumnsFromKeys } from "~/components/dataTable/column";
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

const isCreateOrUpdate = ref<boolean>(false);
const isEditFormDialogOpen = ref<boolean>(false);
const confirmUpdateDialog = ref<boolean>(false);
// const isLoading = ref(false);

// dialog form config
// type is input type (text , number , textarea , file etc)
const editfields = {
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
        console.log(acc, "a");
        console.log(model, "b");
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

// async function createModel() {
//   const formData = new FormData();
//   formData.append("name", modelForm.name);
//   formData.append("description", modelForm.description);
//   if (modelForm.file) {
//     formData.append("file", modelForm.file);
//   }

//   const projectId = route.params.projectId as string;
//   const response = await $fetch<Model>(
//     `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models`,
//     {
//       method: "POST",
//       body: formData,
//     },
//   );

//   models.value[response.id] = response;
// }

// async function updateModel(modelId: string) {
//   const projectId = route.params.projectId as string;

//   try {
//     const response = await $fetch<Model>(
//       `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models/${modelId}`,
//       {
//         method: "PUT",
//         body: {
//           name: modelForm.name,
//           description: modelForm.description,
//         },
//       },
//     );

//     models.value[modelId] = {
//       name: response.name,
//       description: response.description,
//       version: response.version,
//       createdAt: response.createdAt,
//       updatedAt: response.updatedAt,
//       projectId: response.projectId,
//     };

//     console.log("Updated row:", modelId);
//   } catch (err) {
//     console.error("Update failed", err);
//   }
// }

// async function deleteRow(id: string) {
//   try {
//     const projectId = route.params.projectId as string;

//     await $fetch(
//       `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models/${id}`,
//       {
//         method: "DELETE",
//       },
//     );

//     models.value = Object.fromEntries(
//       Object.entries(models.value).filter(([key]) => key !== id),
//     );
//     console.log("Deleted row:", id);
//   } catch (err) {
//     console.error("Delete failed", err);
//   }
// }

function handleEditRow(row: Model) {
  console.log("ad");
  currentEditId.value = row.id;
  isCreateOrUpdate.value = true;
  isEditFormDialogOpen.value = true;
}

function handleDeleteRow(row: Model) {
  currentEditId.value = row.id;
  confirmUpdateDialog.value = true;
}

// Form submit handler
function handleFormSubmit() {
  isEditFormDialogOpen.value = true;
  confirmUpdateDialog.value = true;
}

onMounted(() => {
  fetchModel();
  console.log("abc", generateColumnsFromKeys<Model>(modelKeys, tableTitles));
});

// data computed when change (record -> array)
const dataArray = computed(() =>
  Object.entries(models.value).map(([id, model]) => ({ id, ...model })),
);
</script>

<template>
  <div>
    <div class="container py-10 mx-auto">
      <DataTable :columns="generateKey" :data="dataArray" />
    </div>
    <FormDialog
      v-model:open="isEditFormDialogOpen"
      :fields="editfields"
      :model="modelForm"
      :titles="formTitles"
      @submit="handleFormSubmit"
    />
  </div>
</template>
