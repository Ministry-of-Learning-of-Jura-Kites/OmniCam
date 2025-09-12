<script setup lang="ts">
import { Button } from "@/components/ui/button";
import FormDialog from "~/components/dialog/FormDialog.vue";
import ConfirmDialog from "~/components/dialog/ConfirmDialog.vue"; // นำเข้า ConfirmDialog

interface Project {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

const config = useRuntimeConfig();
const projects = ref<Project[]>([]);
const loading = ref(false);
const error = ref<any>(null);

async function fetchProjects() {
  try {
    loading.value = true;
    error.value = null;
    const { data } = await $fetch<{ data: Project[] }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects`,
      { method: "GET" },
    );
    projects.value = data;
  } catch (err) {
    error.value = err;
    console.error("Error fetching projects:", err);
  } finally {
    loading.value = false;
  }
}

// -------- Dialog State & Form --------
const isFormDialogOpen = ref(false);
const isCreateOrUpdate = ref(false);
const currentEditId = ref<string | null>(null);

// -------- Confirm Dialog State --------
const isConfirmDialogOpen = ref(false);
const confirmAction = ref<"update" | "delete" | null>(null);
const confirmMessage = ref("");

type ProjectForm = { name: string; description: string };
const projectForm = reactive<ProjectForm>({
  name: "",
  description: "",
});

const formFields = {
  name: "text",
  description: "textarea",
} as const;

const formTitles = {
  name: "Project Name",
  description: "Description",
};

async function createProject() {
  try {
    const body = {
      name: projectForm.name,
      description: projectForm.description,
    };
    const { data } = await $fetch<{ data: Project }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects`,
      { method: "POST", body },
    );
    projects.value = [...projects.value, data];
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
    projects.value = projects.value.map((p) => (p.id === id ? data : p));
  } catch (err) {
    console.error("Error updating project:", err);
  }
}

async function deleteProject(id: string) {
  try {
    await $fetch(`${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${id}`, {
      method: "DELETE",
    });
    projects.value = projects.value.filter((p) => p.id !== id);
  } catch (err) {
    console.error("Error deleting project:", err);
  }
}

// -------- Handlers --------
function openCreateDialog() {
  isCreateOrUpdate.value = true;
  currentEditId.value = null;
  projectForm.name = "";
  projectForm.description = "";
  isFormDialogOpen.value = true;
}

function handleEditRow(project: Project) {
  isCreateOrUpdate.value = false;
  currentEditId.value = project.id;
  projectForm.name = project.name;
  projectForm.description = project.description;
  isFormDialogOpen.value = true;
}

async function handleFormSubmit() {
  if (!isCreateOrUpdate.value && currentEditId.value) {
    confirmAction.value = "update";
    confirmMessage.value = `Are you sure you want to update "${projectForm.name}" project?`;
    isFormDialogOpen.value = false;
    isConfirmDialogOpen.value = true;
  } else if (isCreateOrUpdate.value) {
    await createProject();
    isFormDialogOpen.value = false;
  }
}

function handleDeleteProject(id: string, name: string) {
  currentEditId.value = id;
  confirmAction.value = "delete";
  confirmMessage.value = `Are you sure you want to delete "${name}" project?`;
  isConfirmDialogOpen.value = true;
}

async function handleConfirmAction() {
  if (confirmAction.value === "update" && currentEditId.value) {
    await updateProject(currentEditId.value);
    isFormDialogOpen.value = false;
  } else if (confirmAction.value === "delete" && currentEditId.value) {
    await deleteProject(currentEditId.value);
  }

  isConfirmDialogOpen.value = false;
  confirmAction.value = null;
  currentEditId.value = null;
}

onMounted(async () => {
  await fetchProjects();
});
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

    <div v-if="loading && projects?.length === 0" class="text-center py-8">
      <p>Loading projects...</p>
    </div>

    <div v-else class="gap-4">
      <div
        v-for="project in projects"
        :key="project.id"
        class="p-4 rounded-lg shadow mb-4"
      >
        <h3 class="text-xl font-semibold">{{ project.name }}</h3>
        <p class="text-gray-600">{{ project.description }}</p>
        <p class="text-sm text-gray-500">
          Created: {{ new Date(project.createdAt).toLocaleDateString() }}
        </p>

        <div class="mt-4 flex gap-2">
          <Button
            variant="outline"
            @click="$router.push(`/projects/${project.id}`)"
          >
            View Details
          </Button>
          <Button variant="secondary" @click="handleEditRow(project)">
            Edit
          </Button>
          <Button
            variant="destructive"
            @click="handleDeleteProject(project.id, project.name)"
          >
            Delete
          </Button>
        </div>
      </div>

      <div v-if="projects?.length === 0" class="text-center py-8 text-gray-500">
        No projects found. Create your first project!
      </div>
    </div>

    <FormDialog
      v-model:open="isFormDialogOpen"
      v-model:model="projectForm"
      :fields="formFields"
      :titles="formTitles"
      :mode="isCreateOrUpdate ? 'create' : 'update'"
      @submit="handleFormSubmit"
    />

    <ConfirmDialog
      v-model:open="isConfirmDialogOpen"
      :message="confirmMessage"
      @submit="handleConfirmAction"
      @close="isConfirmDialogOpen = false"
    />
  </div>
</template>
