<!-- components/projects/ProjectCreateDialog.vue -->
<script setup lang="ts">
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
  projectCreated: [];
}>();

const name = ref("");
const description = ref("");
const isLoading = ref(false);
const error = ref<string | null>(null);

// Reset form when dialog opens/closes
watch(
  () => props.open,
  (newVal) => {
    if (!newVal) {
      resetForm();
    }
  },
);

const resetForm = () => {
  name.value = "";
  description.value = "";
  error.value = null;
  isLoading.value = false;
};

const createProject = async () => {
  if (!name.value.trim()) {
    error.value = "Project name is required";
    return;
  }

  try {
    isLoading.value = true;
    error.value = null;

    // ใช้ $fetch โดยตรง
    await $fetch("http://localhost:8080/api/v1/projects", {
      method: "POST",
      body: {
        name: name.value,
        description: description.value,
      },
    });

    // Emit event เพื่อให้ parent refresh data
    emit("projectCreated");

    // Close dialog and reset form
    emit("update:open", false);
    resetForm();
  } catch (err) {
    error.value = "Failed to create project";
    console.error("Error creating project:", err);
  } finally {
    isLoading.value = false;
  }
};

const closeDialog = () => {
  emit("update:open", false);
  resetForm();
};
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>Create New Project</DialogTitle>
      </DialogHeader>

      <!-- Error Message -->
      <div
        v-if="error"
        class="bg-destructive/15 text-destructive px-3 py-2 rounded-md text-sm"
      >
        {{ error }}
      </div>

      <div class="grid gap-4 py-4">
        <div class="grid gap-2">
          <Label for="name" class="text-sm font-medium"> Project Name </Label>
          <Input
            id="name"
            v-model="name"
            placeholder="Enter project name"
            :disabled="isLoading"
            class="w-full"
          />
        </div>

        <div class="grid gap-2">
          <Label for="description" class="text-sm font-medium">
            Description
          </Label>
          <Textarea
            id="description"
            v-model="description"
            placeholder="Enter project description (optional)"
            :disabled="isLoading"
            class="min-h-[100px]"
          />
        </div>
      </div>

      <div class="flex justify-end gap-3">
        <Button variant="outline" @click="closeDialog" :disabled="isLoading">
          Cancel
        </Button>
        <Button @click="createProject" :disabled="isLoading">
          <span v-if="isLoading">Creating...</span>
          <span v-else>Create Project</span>
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
