<script setup lang="ts">
// import { keyof } from "zod";
// import { generateColumnsFromKeys } from "~/components/dataTable/column";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue";
import FormDialog from "~/components/dialog/FormDialog.vue";
import SuccessDialog from "~/components/dialog/SuccessDialog.vue";
import ContentCard from "~/components/card/ContentCard.vue";
import CustomPagination from "~/components/pagination/CustomPagination.vue";

export interface Model {
  id: string;
  projectId: string;
  name: string;
  description: string;
  version: number;
  imagePath?: string;
  filePath?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ModelGetRequest {
  data: Model[];
  count: number;
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
  image: File | null;
};
export type ModelWithoutId = Omit<Model, "id">;

const config = useRuntimeConfig();
const route = useRoute();

const models = ref<Record<string, ModelWithoutId>>({});
const totalData = ref<number>(0);
const currentEditId = ref<string | null>(null);
const modelForm = reactive<ModelForm>({
  name: "",
  description: "",
  file: null,
  image: null,
});

//pagination
const page = ref<number>(1);
const pageSize = ref<number>(4);

// write table model (what u want to show etc.)
// const modelKeys: (keyof Model)[] = [
//   "name",
//   "description",
//   "version",
//   "createdAt",
//   "updatedAt",
// ];

// const tableTitles = {
//   name: "Name",
//   description: "Description",
//   version: "Version",
//   createdAt: "Created At",
//   updatedAt: "Updated At",
// };

// generate key to use here

// const generateKey = generateColumnsFromKeys<Model>(modelKeys, tableTitles, {
//   onEdit: (row) => handleEditRow(row),
//   onDelete: (row) => handleDeleteRow(row),
// });

// Dialog handler for create delete update
const isEditFormDialogOpen = ref<boolean>(false);
const isCreateFormDialogOpen = ref<boolean>(false);
const confirmDialog = ref<boolean>(false);
const confirmMessage = ref<string>("");
const successDialog = ref<boolean>(false);
const successMessage = ref<string>("");
// const isLoading = ref(false);

// dialog form config
// type is input type (text , number , textarea , file etc)
const editfields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: true },
  // file: "file",
} as const;

const createFields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: true },
  file: { type: "file" as const, required: true },
  image: { type: "file" as const, required: true },
} as const;

const formTitles = {
  name: "Name",
  description: "Description",
  file: "Model file",
  image: "Model Image",
};

async function fetchModel() {
  const projectId = route.params.projectId as string;
  try {
    const response = await $fetch<ModelGetRequest>(
      `http://http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models`,
      {
        method: "GET",
        query: {
          pageSize: pageSize.value,
          page: page.value,
        },
      },
    );
    const now = Date.now();
    if (response.data != null) {
      models.value = response.data.reduce<Record<string, ModelWithoutId>>(
        (acc, model) => {
          const { id, imagePath, ...rest } = model;
          acc[id] = {
            ...rest,
            imagePath: imagePath ? `${imagePath}?t=${now}` : undefined,
          };
          return acc;
        },
        {},
      );
      totalData.value = response.count;
    } else {
      models.value = {};
      totalData.value = 0;
    }

    console.log(models.value, "c");
    console.log(totalData.value);
    console.log(totalData.value);
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

  if (modelForm.image) {
    formData.append("image", modelForm.image);
  }

  for (const [key, value] of formData.entries()) {
    // For files, log the name
    if (value instanceof File) {
      console.log(`${key}: ${value.name} (${value.size} bytes)`);
    } else {
      console.log(`${key}: ${value}`);
    }
  }

  const projectId = route.params.projectId as string;
  const response = await $fetch<ModelReturnRequest>(
    `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models`,
    {
      method: "POST",
      body: formData,
    },
  );

  successDialog.value = true;
  successMessage.value = `You have successfully created ${response.data.name}`;

  fetchModel();
}

async function updateModel(modelId: string) {
  const projectId = route.params.projectId as string;
  console.log(modelId);
  console.log(modelForm);
  try {
    const response = await $fetch<ModelReturnRequest>(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models/${modelId}`,
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
        imagePath: response.data.imagePath,
        version: response.data.version,
        createdAt: response.data.createdAt,
        updatedAt: response.data.updatedAt,
        projectId: response.data.projectId,
      },
    };
    console.log("new Model", models.value[modelId]);
    successDialog.value = true;
    successMessage.value = `You have successfully update ${response.data.name}`;
    console.log("Updated row:", modelId);
  } catch (err) {
    console.error("Update failed", err);
  }
}

async function deleteRow(id: string) {
  try {
    const projectId = route.params.projectId as string;

    await $fetch(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models/${id}`,
      {
        method: "DELETE",
      },
    );

    models.value = Object.fromEntries(
      Object.entries(models.value).filter(([key]) => key !== id),
    );
    console.log("Deleted row:", id);
    successDialog.value = true;
    successMessage.value = `You have successfully delete ${id}`;

    await fetchModel();
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

async function handleUpdateImage(file: File | undefined, modelId: string) {
  if (!file || !modelId) return;

  const projectId = route.params.projectId;
  const formData = new FormData();
  formData.append("image", file);

  try {
    const res = await $fetch<{ imagePath: string; message: string }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${projectId}/models/${modelId}/image`,
      {
        method: "PUT",
        body: formData,
      },
    );

    const updatedImagePath = res.imagePath;
    const timestamp = new Date().getTime();
    await new Promise((resolve) => setTimeout(resolve, 500));
    // Ensure the model exists and then create a new object
    if (models.value[modelId]) {
      models.value[modelId] = {
        ...models.value[modelId],
        imagePath: `${updatedImagePath}?t=${timestamp}`,
      };
    }

    successDialog.value = true;
    successMessage.value = `Image updated successfully for model ${modelId}`;
  } catch (err) {
    console.error("Failed to update image", err);
  }
}

onMounted(() => {
  fetchModel();
});

watch(
  modelForm,
  (modelForm) => {
    console.log(modelForm);
  },
  {
    once: false,
  },
);

watch([page, pageSize], async ([page, pageSize]) => {
  console.log(page);
  console.log(pageSize);
  await fetchModel();
});
</script>

<template>
  <div>
    <div class="flex flex-col items-center min-h-screen p-4">
      <div class="w-full max-w-7xl flex justify-end mb-4">
        <Button type="button" @click="handleCreate" />
      </div>

      <div class="w-full max-w-7xl flex justify-center mb-4">
        <div class="flex flex-row gap-6 overflow-x-auto w-full">
          <ContentCard
            v-for="(model, id) in models"
            :key="id"
            :name="model.name"
            :description="model.description"
            :image-path="model.imagePath ?? ''"
            :redirect-link="`models/${id}`"
            @update="handleEditRow({ ...model, id })"
            @delete="handleDeleteRow({ ...model, id })"
            @update-image="
              (file: File | undefined) => handleUpdateImage(file, id)
            "
          />
        </div>
      </div>

      <div class="w-full max-w-7xl flex justify-center">
        <CustomPagination
          v-model:page="page"
          :page-size="pageSize"
          :total-item="totalData"
        />
      </div>
    </div>

    <!-- <p>Current Page: {{ page }}</p> -->
    <!-- <div class="container py-10 mx-auto">
      <DataTable :columns="generateKey" :data="dataArray" />
    </div> -->
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
    <SuccessDialog
      v-model:open="successDialog"
      :message="successMessage"
      icon="fa fa-check-circle"
    />
  </div>
</template>
