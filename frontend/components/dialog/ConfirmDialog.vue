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
    <DialogContent>
      <DialogHeader class="flex flex-col items-center space-y-4">
        <h2 class="text-center justify-center flex">{{ message }}</h2>
      </DialogHeader>

      <DialogFooter>
        <DialogClose as-child>
          <Button type="button" variant="secondary" @click="handleClose">
            Cancel
          </Button>
        </DialogClose>
        <Button type="button" @click="handleSubmit">
          {{
            props.message.includes("delete") ? "Yes!, Delete" : "Yes!, Update"
          }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
