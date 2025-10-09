<script setup lang="ts">
// import { keyof } from "zod";
// import { generateColumnsFromKeys } from "~/components/dataTable/column";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue";
import FormDialog from "~/components/dialog/FormDialog.vue";
import SuccessDialog from "~/components/dialog/SuccessDialog.vue";
import ContentCard from "~/components/card/ContentCard.vue";
import CustomPagination from "~/components/pagination/CustomPagination.vue";
import { uuidToBase64Url } from "~/lib/utils";
import { Plus } from "lucide-vue-next";
import type { Project } from "~/types/project";

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
const projectDetail = ref<Project | null>(null);
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

async function fetchProjectById() {
  const projectId = route.params.projectId as string;
  try {
    const response = await $fetch<{ data: Project }>(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}`,
      {
        method: "GET",
      },
    );

    projectDetail.value = response.data;
  } catch (err) {
    console.error("Failed to fetch project by id", err);
  }
}

async function fetchModel() {
  const projectId = route.params.projectId as string;
  try {
    const response = await $fetch<ModelGetRequest>(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models`,
      {
        method: "GET",
        query: {
          pageSize: pageSize.value,
          page: page.value,
        },
        credentials: "include",
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
      credentials: "include",
    },
  );

  successDialog.value = true;
  successMessage.value = `You have successfully created ${response.data.name}`;

  fetchModel();
}

async function updateModel(modelId: string) {
  const projectId = route.params.projectId as string;
  try {
    const response = await $fetch<ModelReturnRequest>(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models/${modelId}`,
      {
        method: "PUT",
        body: {
          name: modelForm.name,
          description: modelForm.description,
        },
        credentials: "include",
      },
    );
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
    successDialog.value = true;
    successMessage.value = `You have successfully update ${response.data.name}`;
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
        credentials: "include",
      },
    );

    models.value = Object.fromEntries(
      Object.entries(models.value).filter(([key]) => key !== id),
    );
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
  isEditFormDialogOpen.value = false;
  confirmDialog.value = true;
}

function handleConfirmSubmit() {
  if (!currentEditId.value) {
    return;
  }

  if (confirmMessage.value.includes("delete")) {
    deleteRow(currentEditId.value);
  } else {
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
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models/${modelId}/image`,
      {
        method: "PUT",
        body: formData,
        credentials: "include",
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
  fetchProjectById();
  fetchModel();
});

watch([page, pageSize], async () => {
  await fetchModel();
});
</script>

<template>
  <div class="flex flex-col min-h-screen p-6">
    <!-- Header -->
    <div
      class="w-full max-w-7xl mx-auto rounded-2xl shadow p-6 mb-6 border border-gray-300"
    >
      <div
        class="flex flex-col md:flex-row md:items-center md:justify-between mb-4"
      >
        <h1 class="text-2xl font-semibold">
          {{ projectDetail?.name }}
        </h1>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <p class="font-medium">Created</p>
          <p>
            {{
              projectDetail
                ? new Date(projectDetail.createdAt).toLocaleDateString()
                : "-"
            }}
          </p>
        </div>
        <div>
          <p class="font-medium">Last Modified</p>
          <p>
            {{
              projectDetail
                ? new Date(projectDetail.updatedAt).toLocaleDateString()
                : "-"
            }}
          </p>
        </div>

        <div>
          <p class="font-medium">Total Models</p>
          <p>{{ totalData }} models</p>
        </div>
        <div>
          <p class="font-medium">Team Members</p>
          <p>5 members</p>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="w-full max-w-7xl mx-auto flex flex-col lg:flex-row gap-6">
      <!-- Left Section: Models -->
      <div class="flex-1 rounded-2xl shadow p-6 border border-gray-300">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold">3D Models</h2>
          <input
            type="text"
            placeholder="Search models by name..."
            class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-60 focus:ring-2 focus:ring-blue-400"
          />
          <Button type="button" class="mt-3 md:mt-0" @click="handleCreate">
            <Plus class="w-4 h-4 mr-1" /> Upload Models
          </Button>
        </div>

        <!-- Models Grid -->
        <div
          class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-6 justify-items-center overflow-y-auto max-h-[750px]"
        >
          <ContentCard
            v-for="(model, id) in models"
            :key="id"
            class="w-full max-w-[280px]"
            :name="model.name"
            :description="model.description"
            :image-path="model.imagePath ?? ''"
            :redirect-link="`/projects/${route.params.projectId}/models/${uuidToBase64Url(id)}`"
            @update="handleEditRow({ ...model, id })"
            @delete="handleDeleteRow({ ...model, id })"
            @update-image="
              (file: File | undefined) => handleUpdateImage(file, id)
            "
          />
        </div>

        <!-- Pagination -->
        <div class="flex justify-center mt-6">
          <CustomPagination
            v-model:page="page"
            :page-size="pageSize"
            :total-item="totalData"
          />
        </div>
      </div>

      <!-- Right Section: Project Info -->
      <div class="w-full lg:w-80 flex-shrink-0 flex flex-col gap-6">
        <div class="rounded-2xl shadow p-5 border border-gray-300">
          <h3 class="text-md font-semibold mb-3">Project Information</h3>
          <div class="space-y-2 text-sm">
            <div>
              <p class="font-medium">Project Name :</p>
              <p>{{ projectDetail?.name }}</p>
            </div>
            <div>
              <p class="font-medium">Description :</p>
              <p>
                {{ projectDetail?.description }}
              </p>
            </div>
          </div>
          <Button class="mt-4 w-full">Edit Information</Button>
        </div>

        <div class="rounded-2xl shadow p-5 border border-gray-300">
          <h3 class="text-md font-semibold mb-3">Team Members</h3>
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-full flex items-center justify-center">
              JD
            </div>
            <div class="text-sm">
              <p class="font-medium">John Doe</p>
              <p class="">Owner</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Dialogs -->
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
