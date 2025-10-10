<script setup lang="ts">
import { ref, watch } from "vue";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";

const props = defineProps<{
  open: boolean;
  currentRole: string;
  roles: string[];
}>();

const emit = defineEmits<{
  (e: "update:open", val: boolean): void;
  (e: "submit", newRole: string): void;
}>();

const selectedRole = ref(props.currentRole);

watch(
  () => props.open,
  (val) => {
    if (val) selectedRole.value = props.currentRole;
  },
);

function handleSubmit() {
  emit("submit", selectedRole.value);
  emit("update:open", false);
}

function handleClose() {
  emit("update:open", false);
}

function formatRoleText(role: string): string {
  const formatMap: Record<string, string> = {
    owner: "Owner",
    project_manager: "Project Manager",
    collaborator: "Collaborator",
  };
  return formatMap[role] || role;
}
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent
      class="max-w-md w-full h-[350px] rounded-2xl p-6 shadow-lg flex flex-col items-center justify-center space-y-6"
    >
      <h3 class="text-lg font-semibold mb-4">Edit Role</h3>

      <Select v-model="selectedRole" class="w-full">
        <SelectTrigger>
          <SelectValue placeholder="Select role" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="role in roles" :key="role" :value="role">
            {{ formatRoleText(role) }}
          </SelectItem>
        </SelectContent>
      </Select>

      <div class="flex justify-end gap-2 w-full mt-4">
        <Button variant="outline" @click="handleClose"> Cancel </Button>
        <Button @click="handleSubmit"> Save </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
