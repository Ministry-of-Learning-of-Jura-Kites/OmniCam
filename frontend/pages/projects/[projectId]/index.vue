<script setup lang="ts">
// import { keyof } from "zod";
// import { generateColumnsFromKeys } from "~/components/dataTable/column";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue";
import SuccessDialog from "~/components/dialog/SuccessDialog.vue";
import FormDialog from "~/components/dialog/FormDialog.vue";
import AddUserDialog from "~/components/dialog/AddUserDialog.vue";
import EditRoleDialog from "~/components/dialog/EditRoleDialog.vue";
import { useAuth } from "~/composables/api/use-auth";
import ContentCard from "~/components/card/ContentCard.vue";
import CustomPagination from "~/components/pagination/CustomPagination.vue";
import { uuidToBase64Url } from "~/lib/uuid";
import { Plus, ArrowLeft } from "lucide-vue-next";
import { useProject, type ProjectMember } from "~/composables/api/use-project";
import {
  type Model,
  useModels,
  getUrlForModelImage,
} from "~/composables/api/use-models";

export type ModelForm = {
  name: string;
  description: string;
  file: File | null;
  image: File | null;
};
export type ModelWithoutId = Omit<Model, "modelId">;

const config = useRuntimeConfig();

const route = useRoute();
const { user, fetchUser } = useAuth();
fetchUser();

const projectId = route.params.projectId as string;

const projectApi = useProject();
const modelApi = useModels(projectId);

const members = ref<ProjectMember[]>([]);
const currentEditId = ref<string | null>(null);
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

// Dialog handler for create delete update
const isEditFormDialogOpen = ref<boolean>(false);
const isCreateFormDialogOpen = ref<boolean>(false);
const isAddUserOpen = ref<boolean>(false);
const isEditRoleOpen = ref(false);
const confirmDialog = ref<boolean>(false);
const confirmMessage = ref<string>("");
const successDialog = ref<boolean>(false);
const successMessage = ref<string>("");

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
  image: { type: "file" as const, required: false },
} as const;

const formTitles = {
  name: "Name",
  description: "Description",
  file: "Model file",
  image: "Model Image",
};
const roles = ["project_manager", "collaborator"];

const { data: project } = await useAsyncData(
  `project-${projectId}`,
  () => projectApi.getProject(projectId),
  {
    // Extract the .data property immediately so 'project' is the object itself
    transform: (response) => response.data,
  },
);
const projectDetail = ref(project.value);

watch(project, (newVal) => {
  projectDetail.value = newVal;
});

// const { data: modelsRaw, refresh } = await useAsyncData(
//   `models-list-${projectId}`,
//   () => modelApi.listModels(page.value, pageSize.value),
//   {
//     watch: [page, pageSize],

//     transform: (respData) => {
//       if (!respData?.data) return { record: {}, count: 0 };
//       const record = respData.data.reduce<Record<string, ModelWithoutId>>(
//         (acc, model) => {
//           const { modelId, imagePath, imageExtension, ...rest } = model;

//           let urlHref: string | undefined = undefined;
//           if (imageExtension != null && imagePath) {
//             urlHref = getUrlForModelImage(
//               projectId,
//               modelId,
//               imageExtension,
//             ).href;
//           }

//           acc[modelId] = {
//             ...rest,
//             imagePath: urlHref,
//           };
//           console.log("record", record);
//           return acc;
//         },
//         {},
//       );
//       return { record, count: respData.count };
//     },
//   },
// );

const asyncKey = computed(
  () => `models-list-${projectId}-page-${page.value}-size-${pageSize.value}`,
);

const { data: modelsRaw, refresh } = await useAsyncData(
  asyncKey.value,
  () => modelApi.listModels(page.value, pageSize.value),
  {
    // server: false, // Disable server-side fetching to ensure it only runs on the client,
    watch: [page, pageSize],
    default: () => ({ record: {} as Record<string, ModelWithoutId>, count: 0 }),
    transform: (respData) => {
      console.log("transform called with:", JSON.stringify(respData));
      if (!respData?.data) return { record: {}, count: 0 };
      const record = respData.data.reduce<Record<string, ModelWithoutId>>(
        (acc, model) => {
          const { modelId, imagePath, imageExtension, ...rest } = model;
          let urlHref: string | undefined = undefined;
          if (imageExtension != null && imagePath) {
            urlHref = getUrlForModelImage(
              projectId,
              modelId,
              imageExtension,
            ).href;
          }
          acc[modelId] = { ...rest, imagePath: urlHref };
          return acc;
        },
        {},
      );
      return { record, count: respData.count };
    },
  },
);

// console.log("modelsRaw", modelsRaw.value);
const models = computed(() => modelsRaw.value?.record ?? {});
const totalData = computed(() => modelsRaw.value?.count ?? 0);

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

  const response = await modelApi.postCreateModel(formData);

  successDialog.value = true;
  successMessage.value = `You have successfully created ${response.data.name}`;

  await refresh();
}

async function updateModel(hexId: string) {
  const modelId = uuidToBase64Url(hexId);
  try {
    const response = await modelApi.updateModel(modelId, {
      name: modelForm.name,
      description: modelForm.description,
    });
    models.value = {
      ...models.value,
      [hexId]: {
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
    const modelId = uuidToBase64Url(id);
    await modelApi.deleteModel(modelId);

    models.value = Object.fromEntries(
      Object.entries(models.value).filter(([key]) => key !== id),
    );
    successDialog.value = true;
    successMessage.value = `You have successfully delete ${id}`;

    await refresh();
  } catch (err) {
    console.error("Delete failed", err);
  }
}
async function fetchMembers() {
  try {
    const res = await projectApi.getProjectMembers(projectId);
    members.value = res.data;
  } catch (err) {
    console.error("Failed to fetch members", err);
  }
}

async function deleteMember(userId: string) {
  try {
    const projectId = route.params.projectId as string;
    const encodedUserId = uuidToBase64Url(userId);
    await projectApi.removeProjectMember(projectId, encodedUserId);

    successDialog.value = true;
    successMessage.value = `Member removed successfully`;

    await fetchMembers();
  } catch (err) {
    console.error("Delete member failed", err);
  }
}

async function handleSubmitRole(newRole: string) {
  if (!editingMember.value) return;
  const encodedUserId = uuidToBase64Url(editingMember.value.userId);

  await projectApi.updateProjectMember(projectId, encodedUserId, {
    role: newRole,
  });
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
  const formData = new FormData();
  formData.append("image", file);

  try {
    const res = await modelApi.updateModelImage(modelId, formData);

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

function goToProjects() {
  navigateTo(
    `http://${config.public.nuxtHost}:${config.public.nuxtPort}/projects`,
    { external: true },
  );
}

fetchMembers();
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
        <div class="flex items-center gap-3">
          <Button
            class="cursor-pointer hover:bg-gray-100"
            variant="ghost"
            size="sm"
            @click="goToProjects()"
          >
            <ArrowLeft class="w-4 h-4" />
          </Button>
          <h1 class="text-2xl font-semibold">{{ projectDetail?.name }}</h1>
        </div>
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
      <div class="w-full lg:w-80 shrink-0 flex flex-col gap-6">
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
