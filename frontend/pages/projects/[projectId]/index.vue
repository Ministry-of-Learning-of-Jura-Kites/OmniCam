<script setup lang="ts">
// import { keyof } from "zod";
// import { generateColumnsFromKeys } from "~/components/dataTable/column";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue";
import SuccessDialog from "~/components/dialog/SuccessDialog.vue";
import FormDialog from "~/components/dialog/FormDialog.vue";
import AddUserDialog from "~/components/dialog/AddUserDialog.vue";
import EditRoleDialog from "~/components/dialog/EditRoleDialog.vue";
import { useAuth } from "~/composables/useAuth";
import ContentCard from "~/components/card/ContentCard.vue";
import CustomPagination from "~/components/pagination/CustomPagination.vue";
import { uuidToBase64Url } from "~/lib/utils";
import { Plus } from "lucide-vue-next";
import type { Project } from "~/types/project";

export interface Model {
  modelId: string;
  projectId: string;
  name: string;
  description: string;
  version: number;
  imagePath?: string;
  filePath?: string;
  createdAt: string;
  updatedAt: string;
}
export interface ProjectMember {
  userId: string;
  username: string;
  firstName?: string | null;
  lastName?: string | null;
  role: string;
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
export type ModelWithoutId = Omit<Model, "modelId">;

const config = useRuntimeConfig();
const route = useRoute();
const auth = await useAuth();
const { user } = auth;
console.log("user in project page", user);

const models = ref<Record<string, ModelWithoutId>>({});
const members = ref<ProjectMember[]>([]);
const totalData = ref<number>(0);
const currentEditId = ref<string | null>(null);
const projectDetail = ref<Project | null>(null);
const editingMember = ref<ProjectMember | null>(null);
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
const isAddUserOpen = ref<boolean>(false);
const isEditRoleOpen = ref(false);
const confirmDialog = ref<boolean>(false);
const confirmMessage = ref<string>("");
const successDialog = ref<boolean>(false);
const successMessage = ref<string>("");
console.log("member", user.value);
const userProjectRole = computed<
  "owner" | "project_manager" | "collaborator" | null
>(() => {
  if (!user.value || members.value.length === 0) return null;
  const me = members.value.find((m) => m.username === user.value?.username);
  return me?.role as "owner" | "project_manager" | "collaborator" | null;
});

// const isLoading = ref(false);

// dialog form config
// type is input type (text , number , textarea , file etc)
const editfields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: false },
  // file: "file",
} as const;

const createFields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: false },
  file: { type: "file" as const, required: true },
  image: { type: "file" as const, required: true },
} as const;

const formTitles = {
  name: "Name",
  description: "Description",
  file: "Model file",
  image: "Model Image",
};
const roles = ["project_manager", "collaborator"];

async function fetchProjectById() {
  const projectId = route.params.projectId as string;
  try {
    const response = await $fetch<{ data: Project }>(
      `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}`,
      {
        method: "GET",
        credentials: "include",
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
      `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/models`,
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
          const { modelId, imagePath, ...rest } = model;
          acc[modelId] = {
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

  const projectId = route.params.projectId as string;
  const response = await $fetch<ModelReturnRequest>(
    `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/models`,
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

async function updateModel(id: string) {
  const projectId = route.params.projectId as string;
  const modelId = uuidToBase64Url(id);
  try {
    const response = await $fetch<ModelReturnRequest>(
      `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/models/${modelId}`,
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
      [id]: {
        name: response.data.name,
        description: response.data.description,
        imagePath: response.data.imagePath,
        version: response.data.version,
        createdAt: response.data.createdAt,
        updatedAt: response.data.updatedAt,
        projectId: response.data.projectId,
      },
    };
    console.log("models : ", models.value);
    successDialog.value = true;
    successMessage.value = `You have successfully update ${response.data.name}`;
  } catch (err) {
    console.error("Update failed", err);
  }
}

async function deleteRow(id: string) {
  try {
    const projectId = route.params.projectId;
    const modelId = uuidToBase64Url(id);
    await $fetch(
      `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/models/${modelId}`,
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
async function fetchMembers() {
  const projectId = route.params.projectId as string;
  try {
    const res = await $fetch<{ data: ProjectMember[]; count: number }>(
      `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/members`,
      {
        method: "GET",
        credentials: "include",
      },
    );
    members.value = res.data;
  } catch (err) {
    console.error("Failed to fetch members", err);
  }
}

async function deleteMember(userId: string) {
  try {
    const projectId = route.params.projectId as string;
    const encodedUserId = uuidToBase64Url(userId);
    await $fetch(
      `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/member/${encodedUserId}`,
      {
        method: "DELETE",
        credentials: "include",
      },
    );

    successDialog.value = true;
    successMessage.value = `Member removed successfully`;

    await fetchMembers();
  } catch (err) {
    console.error("Delete member failed", err);
  }
}

async function handleSubmitRole(newRole: string) {
  if (!editingMember.value) return;
  const projectId = route.params.projectId;
  const encodedUserId = uuidToBase64Url(editingMember.value.userId);

  await $fetch(
    `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/user/${encodedUserId}/role`,
    {
      method: "PUT",
      body: { role: newRole },
      credentials: "include",
    },
  );
  successDialog.value = true;
  successMessage.value = `${editingMember.value.username}'s role updated to ${newRole}`;
  fetchMembers();
}

function handleDeleteMember(username: string, userId: string) {
  currentEditId.value = userId;
  confirmMessage.value = `Do you want to remove ${username} from this project?`;
  confirmDialog.value = true;
}

function handleEditMember(member: ProjectMember) {
  editingMember.value = member;
  isEditRoleOpen.value = true;
}

function handleCreate() {
  modelForm.name = "";
  modelForm.description = "";
  modelForm.file = null;
  isCreateFormDialogOpen.value = true;
}

function handleEditRow(row: Model) {
  currentEditId.value = row.modelId;
  confirmMessage.value = `Do you want to update this ${row.name}`;
  isEditFormDialogOpen.value = true;
  modelForm.description = row.description;
  modelForm.name = row.name;
  modelForm.file = null;
}

function handleDeleteRow(row: Model) {
  currentEditId.value = row.modelId;
  confirmDialog.value = true;
  confirmMessage.value = `Do you want to delete this ${row.name}`;
}

// Form submit handler
function handleAddUsers() {
  isAddUserOpen.value = true;
}
function handleMembersAdded() {
  fetchMembers();
}
function handleEditFormSubmit() {
  isEditFormDialogOpen.value = false;
  confirmDialog.value = true;
}

function handleConfirmSubmit() {
  if (!currentEditId.value) return;

  if (confirmMessage.value.includes("remove")) {
    deleteMember(currentEditId.value);
  } else if (confirmMessage.value.includes("delete")) {
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

  const projectId = route.params.projectId as string;
  const formData = new FormData();
  formData.append("image", file);

  try {
    const res = await $fetch<{ imagePath: string; message: string }>(
      `http://${config.public.externalBackendHost}/api/v1/projects/${projectId}/models/${uuidToBase64Url(modelId)}/image`,
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
  fetchMembers();
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
          <p>{{ members.length }}</p>
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
            @update="handleEditRow({ ...model, modelId: id })"
            @delete="handleDeleteRow({ ...model, modelId: id })"
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
        </div>

        <div class="rounded-2xl shadow p-5 border border-gray-300">
          <h3 class="text-md font-semibold mb-3">Team Members</h3>

          <div v-if="members.length === 0" class="text-sm text-gray-500 mb-2">
            No members yet
          </div>

          <div
            v-for="m in members"
            :key="m.userId"
            class="flex items-center gap-3 mb-3"
          >
            <div
              class="w-8 h-8 rounded-full bg-gray-200 text-gray-500 flex items-center justify-center uppercase"
            >
              {{ m.username[0] }}
            </div>
            <div class="text-sm flex-1">
              <p class="font-medium">
                {{ m.username }}
                <span v-if="m.username === user?.username" class="text-gray-500"
                  >(me)</span
                >
              </p>
              <p class="text-xs text-gray-500">
                {{ m.role.replace("_", " ") }}
              </p>
            </div>

            <div class="flex gap-1">
              <Button
                size="sm"
                :disabled="
                  m.username === user?.username || userProjectRole !== 'owner'
                "
                @click="handleEditMember(m)"
              >
                Edit
              </Button>
              <Button
                variant="destructive"
                size="sm"
                :disabled="
                  m.username === user?.username ||
                  !(
                    userProjectRole === 'owner' ||
                    (userProjectRole === 'project_manager' &&
                      m.role === 'collaborator')
                  )
                "
                @click="handleDeleteMember(m.username, m.userId)"
              >
                Delete
              </Button>
            </div>
          </div>

          <Button
            class="mt-4 w-full"
            :disabled="
              userProjectRole !== 'owner' &&
              userProjectRole !== 'project_manager'
            "
            @click="handleAddUsers"
          >
            Add Team Members
          </Button>
        </div>
      </div>
    </div>

    <!-- Dialogs -->
    <AddUserDialog
      v-model:open="isAddUserOpen"
      :project-id="route.params.projectId as string"
      :user-role="userProjectRole"
      @submit="handleAddUsers"
      @members-added="handleMembersAdded"
    />

    <EditRoleDialog
      v-model:open="isEditRoleOpen"
      :current-role="editingMember?.role ?? ''"
      :roles="roles"
      @submit="handleSubmitRole"
    />

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
