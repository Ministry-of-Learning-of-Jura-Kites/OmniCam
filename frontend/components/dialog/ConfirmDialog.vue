<script setup lang="ts">
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogFooter,
} from "@/components/ui/dialog";
const props = defineProps<{
  open: boolean;
  message: string;
}>();

const emit = defineEmits<{
  (e: "close" | "submit"): void;
  (e: "update:open", value: boolean): void;
}>();

function handleClose() {
  emit("close");
}

function handleSubmit() {
  emit("submit");
}
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent class="w-fit">
      <DialogHeader class="flex flex-col items-center space-y-4">
        <h2 class="p-3 text-center justify-center flex">{{ message }}</h2>
      </DialogHeader>

      <DialogFooter>
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
