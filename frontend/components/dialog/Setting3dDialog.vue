<script setup lang="ts">
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useSensitivity } from "@/composables/3d-settings";

const props = defineProps<{
  open: boolean;
  message: string;
}>();

const emit = defineEmits<{
  (e: "close" | "submit"): void;
  (e: "update:open", value: boolean): void;
}>();

const { sensitivity, setMouse, setMovement } = useSensitivity();

// Local copies for dialog
const localMouse = ref(sensitivity.value.mouse);
const localMovement = ref(sensitivity.value.movement);

// Reset local values when dialog opens
watch(
  () => props.open,
  (val) => {
    if (val) {
      localMouse.value = sensitivity.value.mouse;
      localMovement.value = sensitivity.value.movement;
    }
  },
);

function handleClose() {
  emit("update:open", false); // close dialog
  emit("close");
}

function handleSubmit() {
  setMouse(localMouse.value);
  setMovement(localMovement.value);
  emit("update:open", false); // close dialog
  emit("submit");
}
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent class="w-fit min-w-[300px]">
      <DialogHeader class="flex flex-col items-center space-y-4">
        <h2 class="p-3 text-center flex justify-center">{{ message }}</h2>

        <!-- Mouse Sensitivity -->
        <div class="w-full space-y-2">
          <label class="text-sm font-medium text-center block">
            Mouse Sensitivity: {{ localMouse }}
          </label>
          <input
            v-model="localMouse"
            type="range"
            min="1"
            max="100"
            class="w-full"
          />
        </div>

        <!-- Movement Sensitivity -->
        <div class="w-full space-y-2">
          <label class="text-sm font-medium text-center block">
            Movement Sensitivity: {{ localMovement }}
          </label>
          <input
            v-model="localMovement"
            type="range"
            min="1"
            max="100"
            class="w-full"
          />
        </div>
      </DialogHeader>

      <DialogFooter class="flex justify-end gap-2">
        <DialogClose as-child>
          <Button type="button" variant="secondary" @click="handleClose">
            Cancel
          </Button>
        </DialogClose>

        <Button type="button" @click="handleSubmit">
          {{
            props.message.toLowerCase().includes("delete") ||
            props.message.toLowerCase().includes("remove")
              ? "Yes!, Delete"
              : "Yes!, Update"
          }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
