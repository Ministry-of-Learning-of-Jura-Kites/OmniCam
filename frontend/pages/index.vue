<script setup lang="ts">
import { Button } from "@/components/ui/button";
import FormDialog from "~/components/dialog/FormDialog.vue";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue";
import SuccessDialog from "~/components/dialog/SuccessDialog.vue";
import ContentCard from "~/components/card/ContentCard.vue";
import CustomPagination from "~/components/pagination/CustomPagination.vue";
import { uuidToBase64Url } from "~/lib/uuid";
import {
  getUrlForProjectImage,
  useProject,
  type Project,
} from "~/composables/api/use-project";

type ProjectWithoutId = Omit<Project, "id">;
type ProjectForm = { name: string; description: string; image: File | null };

const formFields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: false },
  image: { type: "file" as const, required: false },
};
const editFields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: false },
};

const formTitles = {
  name: "Project Name",
  description: "Description",
  image: "Project Image",
};

const projects = ref<Record<string, ProjectWithoutId>>({});
const totalItem = ref(0);
const page = ref(1);
const pageSize = ref(4);

const loading = ref(false);
const error = ref<string | null>(null);

// dialogs & forms
const isFormDialogOpen = ref(false);
const isCreateMode = ref(true);
const currentEditHexId = ref<string | null>(null);

const isConfirmDialogOpen = ref(false);
const confirmAction = ref<"update" | "delete" | null>(null);
const confirmMessage = ref("");

const isSuccessDialogOpen = ref(false);
const successMessage = ref("");

const projectForm = reactive<ProjectForm>({
  name: "",
  description: "",
  image: null,
});

const projectApi = useProject();

const {
  data: respData,
  error: fetchError,
  refresh,
} = await useAsyncData(
  "projects-list", // A unique key ensures the client finds the server's work
  () => projectApi.listProjects(page.value, pageSize.value),
  {
    watch: [page, pageSize],
  },
);

watch(
  respData,
  (newData) => {
    loading.value = false;
    if (fetchError.value != undefined) {
      console.error("Error while fetching projects", fetchError.value);
    }
    const data = newData?.data || [];
    const count = newData?.count || 0;

    projects.value = data.reduce<Record<string, ProjectWithoutId>>((acc, p) => {
      const { id, imagePath, ...rest } = p;

      let url: string | undefined = undefined;
      if (imagePath) {
        url = getUrlForProjectImage(id, imagePath).href;
      }

      acc[id] = {
        ...rest,
        imagePath: url,
      };

      return acc;
    }, {});

    totalItem.value = count;
  },
  { immediate: true },
);

async function submitCreateProject() {
  try {
    const formData = new FormData();
    formData.append("name", projectForm.name);
    formData.append("description", projectForm.description);
    if (projectForm.image) formData.append("image", projectForm.image);

    const { data } = await projectApi.createProject(formData);

    const { id, ...rest } = data;
    projects.value = { [id]: rest, ...projects.value }; // unshift
    successMessage.value = `Project "${data.name}" created successfully.`;
    isSuccessDialogOpen.value = true;
    await refresh();
  } catch (err) {
    console.error("Error creating project:", err);
  }
}

async function submitUpdateProject(hexId: string) {
  try {
    const body = {
      name: projectForm.name,
      description: projectForm.description,
    };

    const { data } = await projectApi.updateProject(hexId, body);
    const { id: pid } = data;

    projects.value[pid] = {
      ...projects.value[pid],
      ...data,
      imagePath: projects.value[pid]?.imagePath,
    };

    successMessage.value = `Project "${data.name}" updated successfully.`;
    isSuccessDialogOpen.value = true;

    await refresh();
  } catch (err) {
    console.error("Error updating project:", err);
  }
}

async function updateProjectImage(id: string, file: File) {
  if (!file) return;

  const formData = new FormData();
  formData.append("image", file);

  try {
    const { imagePath, fileExtension } = await projectApi.updateProjectImage(
      id,
      formData,
    );

    if (projects.value[id]) {
      const imagePathWithExt = `${imagePath}${fileExtension}`;
      const newImageUrl = getUrlForProjectImage(id, imagePathWithExt).href;

      projects.value[id] = {
        ...projects.value[id],
        imagePath: newImageUrl,
      };
    }

    successMessage.value = `Image for project updated successfully.`;
    isSuccessDialogOpen.value = true;
  } catch (err) {
    console.error("Error updating project image:", err);
  }
}

async function deleteProject(hexId: string) {
  try {
    await projectApi.deleteProject(hexId);
    const { [hexId]: _, ...rest } = projects.value;
    projects.value = rest;
    successMessage.value = `Project deleted successfully.`;
    isSuccessDialogOpen.value = true;
    await refresh();
  } catch (err) {
    console.error("Error deleting project:", err);
  }
}

// ---------- Dialog Handlers ----------
function openCreateDialog() {
  isCreateMode.value = true;
  currentEditHexId.value = null;
  projectForm.name = "";
  projectForm.description = "";
  projectForm.image = null;
  isFormDialogOpen.value = true;
}

function handleEditRow(projectHexId: string) {
  isCreateMode.value = false;
  currentEditHexId.value = projectHexId;
  const project = projects.value[projectHexId];
  if (project) {
    projectForm.name = project.name;
    projectForm.description = project.description;
    projectForm.image = null;
    isFormDialogOpen.value = true;
  }
}

async function handleFormSubmit() {
  if (isCreateMode.value) {
    await submitCreateProject();
    isFormDialogOpen.value = false;
  } else if (currentEditHexId.value) {
    confirmAction.value = "update";
    confirmMessage.value = `Update project "${projectForm.name}"?`;
    isFormDialogOpen.value = false;
    isConfirmDialogOpen.value = true;
  }
}

function handleDeleteProject(hexId: string, name: string) {
  currentEditHexId.value = hexId;
  confirmAction.value = "delete";
  confirmMessage.value = `Do you want to delete project "${name}"?`;
  isConfirmDialogOpen.value = true;
}

async function handleConfirmAction() {
  if (confirmAction.value === "update" && currentEditHexId.value) {
    await submitUpdateProject(currentEditHexId.value);
  } else if (confirmAction.value === "delete" && currentEditHexId.value) {
    await deleteProject(currentEditHexId.value);
  }
  isConfirmDialogOpen.value = false;
  confirmAction.value = null;
  currentEditHexId.value = null;
}
</script>

<template>
  <div class="container mx-auto p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">My Projects</h1>
      <Button @click="openCreateDialog">New Project</Button>
    </div>

    <div
      v-if="error"
      class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4"
    >
      {{ error || "Failed to fetch projects" }}
    </div>

    <div
      v-if="Object.keys(projects).length === 0 && !loading"
      class="text-center py-8 text-gray-500"
    >
      No projects found. Create your first project!
    </div>

    <div v-else class="flex flex-row gap-6 overflow-x-auto w-full">
      <ContentCard
        v-for="(project, id) in projects"
        :key="id"
        :name="project.name"
        :description="project.description"
        :redirect-link="`/projects/${uuidToBase64Url(id)}`"
        :image-path="project.imagePath || ''"
        @update="handleEditRow(id)"
        @delete="handleDeleteProject(id, project.name)"
        @update-image="
          (file: File | undefined) => file && updateProjectImage(id, file)
        "
      />
    </div>

    <div class="mt-6 flex justify-center">
      <CustomPagination
        v-model:page="page"
        :page-size="pageSize"
        :total-item="totalItem"
      />
    </div>

    <FormDialog
      v-model:open="isFormDialogOpen"
      v-model:model="projectForm"
      :fields="isCreateMode ? formFields : editFields"
      :titles="formTitles"
      :mode="isCreateMode ? 'create' : 'update'"
      @submit="handleFormSubmit"
    />

    <ConfirmDialog
      v-model:open="isConfirmDialogOpen"
      :message="confirmMessage"
      @submit="handleConfirmAction"
      @close="isConfirmDialogOpen = false"
    />

    <SuccessDialog
      v-model:open="isSuccessDialogOpen"
      :message="successMessage"
      icon="fa fa-check-circle"
    />
  </div>
</template>
