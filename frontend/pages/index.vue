<script setup lang="ts">
import { Button } from "@/components/ui/button";
import ProjectCreateDialog from "~/components/shared/ProjectCreateDialog.vue";

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

const fetchProjects = async () => {
  try {
    loading.value = true;
    error.value = null;

    const { data } = await $fetch<{ data: Project[] }>(
      `${config.public.NUXT_PUBLIC_URL}/api/v1/projects`,
      {
        method: "GET",
      },
    );

    projects.value = data;
  } catch (err) {
    error.value = err;
    console.error("Error fetching projects:", err);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchProjects();
});

const refresh = () => {
  fetchProjects();
};

const isCreateDialogOpen = ref(false);

const deleteProject = async (id: string) => {
  try {
    await $fetch(`${config.public.NUXT_PUBLIC_URL}/api/v1/projects/${id}`, {
      method: "DELETE",
    });
    refresh();
  } catch (err) {
    console.error("Error deleting project:", err);
  }
};

const handleProjectCreated = () => {
  refresh();
};

// Open create dialog
const openCreateDialog = () => {
  isCreateDialogOpen.value = true;
};

console.log(projects, "test");
</script>

<template>
  <div class="container mx-auto p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold">My Projects</h1>
      <Button @click="openCreateDialog"> New Project </Button>
    </div>

    <div
      v-if="error"
      class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4"
    >
      {{ error.message || "Failed to fetch projects" }}
    </div>

    <!-- Loading State -->
    <div v-if="loading && projects?.length === 0" class="text-center py-8">
      <p>Loading projects...</p>
    </div>

    <!-- Projects List -->
    <div v-else class="grid gap-4">
      <div
        v-for="project in projects"
        :key="project.id"
        class="p-4 rounded-lg shadow"
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
          <Button variant="destructive" @click="deleteProject(project.id)">
            Delete
          </Button>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="projects?.length === 0" class="text-center py-8 text-gray-500">
        No projects found. Create your first project!
      </div>
    </div>

    <!-- Create Project Dialog -->
    <ProjectCreateDialog
      v-model:open="isCreateDialogOpen"
      @projectCreated="handleProjectCreated"
    />
  </div>
</template>
