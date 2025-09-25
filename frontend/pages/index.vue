<script setup lang="ts">
import { Button } from "@/components/ui/button";
import FormDialog from "~/components/dialog/FormDialog.vue";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue";
import SuccessDialog from "~/components/dialog/SuccessDialog.vue";
import ContentCard from "~/components/card/ContentCard.vue";
import CustomPagination from "~/components/pagination/CustomPagination.vue";
import type { FieldOption } from "~/components/dialog/types";

interface Project {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
  imagePath?: string;
}
type ProjectWithoutId = Omit<Project, "id">;
type ProjectForm = { name: string; description: string; image: File | null };

const formFields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: true },
  image: { type: "file" as const, required: false },
};
const editFields = {
  name: { type: "text" as const, required: true },
  description: { type: "textarea" as const, required: true },
};

const formTitles = {
  name: "Project Name",
  description: "Description",
  image: "Project Image",
};

const config = useRuntimeConfig();

const projects = ref<Record<string, ProjectWithoutId>>({});
const totalItem = ref(0);
const page = ref(1);
const pageSize = ref(4);

const loading = ref(false);
const error = ref<any>(null);

// dialogs & forms
const isFormDialogOpen = ref(false);
const isCreateMode = ref(true);
const currentEditId = ref<string | null>(null);

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

const dataArray = computed(() =>
  Object.entries(projects.value).map(([id, project]) => ({ id, ...project })),
);

async function fetchProjects() {
  try {
    loading.value = true;
    error.value = null;

    const response = await $fetch<{ data: Project[]; count: number }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects`,
      {
        method: "GET",
        query: { page: page.value, pageSize: pageSize.value },
      },
    );
    const data = response?.data || [];
    const count = response?.count || 0;
    const now = Date.now();
    projects.value = data.reduce<Record<string, ProjectWithoutId>>((acc, p) => {
      const { id, imagePath, ...rest } = p;
      acc[id] = {
        ...rest,
        imagePath: imagePath ? `${imagePath}?t=${now}` : undefined,
      };
      return acc;
    }, {});

    totalItem.value = count;
  } catch (err) {
    error.value = err;
    console.error("Error fetching projects:", err);
  } finally {
    loading.value = false;
  }
}

async function createProject() {
  try {
    const formData = new FormData();
    formData.append("name", projectForm.name);
    formData.append("description", projectForm.description);
    if (projectForm.image) formData.append("image", projectForm.image);

    const { data } = await $fetch<{ data: Project }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects`,
      { method: "POST", body: formData },
    );

    const { id, ...rest } = data;
    projects.value = { [id]: rest, ...projects.value }; // unshift
    successMessage.value = `Project "${data.name}" created successfully.`;
    isSuccessDialogOpen.value = true;
    await fetchProjects();
  } catch (err) {
    console.error("Error creating project:", err);
  }
}

async function updateProject(id: string) {
  try {
    const body = {
      name: projectForm.name,
      description: projectForm.description,
    };
    const { data } = await $fetch<{ data: Project }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${id}`,
      { method: "PUT", body },
    );
    const { id: pid, ...rest } = data;
    projects.value[pid] = rest;
    successMessage.value = `Project "${data.name}" updated successfully.`;
    isSuccessDialogOpen.value = true;
  } catch (err) {
    console.error("Error updating project:", err);
  }
}

async function updateProjectImage(id: string, file: File) {
  if (!file) return;
  const formData = new FormData();
  formData.append("image", file);
  try {
    const { imagePath } = await $fetch<{ imagePath: string }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${id}/image`,
      { method: "PUT", body: formData },
    );
    if (projects.value[id]) {
      projects.value[id] = {
        ...projects.value[id],
        imagePath: `${imagePath}?t=${Date.now()}`,
      };
    }
    successMessage.value = `Image for project updated successfully.`;
    isSuccessDialogOpen.value = true;
  } catch (err) {
    console.error("Error updating project image:", err);
  }
}

async function deleteProject(id: string) {
  try {
    await $fetch(`${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${id}`, {
      method: "DELETE",
    });
    const { [id]: _, ...rest } = projects.value;
    projects.value = rest;
    successMessage.value = `Project deleted successfully.`;
    isSuccessDialogOpen.value = true;
    await fetchProjects();
  } catch (err) {
    console.error("Error deleting project:", err);
  }
}

// ---------- Dialog Handlers ----------
function openCreateDialog() {
  isCreateMode.value = true;
  currentEditId.value = null;
  projectForm.name = "";
  projectForm.description = "";
  projectForm.image = null;
  isFormDialogOpen.value = true;
}

function handleEditRow(projectId: string) {
  isCreateMode.value = false;
  currentEditId.value = projectId;
  const project = projects.value[projectId];
  if (project) {
    projectForm.name = project.name;
    projectForm.description = project.description;
    projectForm.image = null;
    isFormDialogOpen.value = true;
  }
}

async function handleFormSubmit() {
  if (isCreateMode.value) {
    await createProject();
    isFormDialogOpen.value = false;
  } else if (currentEditId.value) {
    confirmAction.value = "update";
    confirmMessage.value = `Update project "${projectForm.name}"?`;
    isFormDialogOpen.value = false;
    isConfirmDialogOpen.value = true;
  }
}

function handleDeleteProject(id: string, name: string) {
  currentEditId.value = id;
  confirmAction.value = "delete";
  confirmMessage.value = `Do you want to delete project "${name}"?`;
  isConfirmDialogOpen.value = true;
}

async function handleConfirmAction() {
  if (confirmAction.value === "update" && currentEditId.value) {
    await updateProject(currentEditId.value);
  } else if (confirmAction.value === "delete" && currentEditId.value) {
    await deleteProject(currentEditId.value);
  }
  isConfirmDialogOpen.value = false;
  confirmAction.value = null;
  currentEditId.value = null;
}

watch([page, pageSize], fetchProjects);
onMounted(fetchProjects);
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
      {{ error.message || "Failed to fetch projects" }}
    </div>

    <div v-if="loading && dataArray.length === 0" class="text-center py-8">
      <p>Loading projects...</p>
    </div>

    <div
      v-else-if="dataArray.length === 0 && !loading"
      class="text-center py-8 text-gray-500"
    >
      No projects found. Create your first project!
    </div>

    <div v-else class="flex flex-row gap-6 overflow-x-auto w-full">
      <ContentCard
        v-for="p in dataArray"
        :key="p.id"
        :name="p.name"
        :description="p.description"
        :redirect-link="`/projects/${p.id}`"
        :image-path="p.imagePath || ''"
        @update="handleEditRow(p.id)"
        @delete="handleDeleteProject(p.id, p.name)"
        @update-image="
          (file: File | undefined) => file && updateProjectImage(p.id, file)
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
