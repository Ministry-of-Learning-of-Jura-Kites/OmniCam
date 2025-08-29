<script setup lang="ts" generic="T extends ModelWithoutId">
import { reactive, watch, defineProps, defineEmits } from "vue";
import { Button } from "@/components/ui/button";
import type { FieldConfig } from "./types";
import type { ModelWithoutId } from "~/pages/projects/[projectId]/index.vue";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

const props = defineProps<{
  model?: Record<string, T>;
  open: boolean;
  fields: FieldConfig[];
  isUpdate?: boolean;
}>();

const emit = defineEmits<{
  (e: "close" | "submit"): void;
  (e: "update:open", value: boolean): void;
}>();

// ✅ form remains generic
const form = reactive<Record<string, T>>({});

// populate form if editing
watch(
  () => props.open,
  (val) => {
    if (val && props.model) {
      Object.assign(form, props.model);
    }
  },
  { immediate: true },
);

function handleSubmit() {
  emit("submit");
}

function handleClose() {
  emit("close");
}
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>
          {{ props.isUpdate ? "Update" : "Create" }}
        </DialogTitle>
      </DialogHeader>

      <div class="flex flex-col gap-2 py-4">
        <template v-for="field in props.fields" :key="field.key">
          <h2>{{ field.key }}</h2>
          <input
            v-if="field.type === 'text' || field.type === 'number'"
            v-model="form[field.key] as unknown as string"
            :type="field.type"
            :placeholder="field.key"
            class="border px-2 py-1 rounded"
          />
          <textarea
            v-else-if="field.type === 'textarea'"
            v-model="form[field.key] as unknown as string"
            :placeholder="field.key"
            class="border px-2 py-1 rounded"
          />
          <input
            v-else-if="field.type === 'file'"
            type="file"
            @change="
              (e) =>
                (form[field.key] = ((e.target as HTMLInputElement).files?.[0] ??
                  null) as unknown as T)
            "
          />
        </template>
      </div>

      <DialogFooter>
        <DialogClose as-child>
          <Button type="button" variant="secondary" @click="handleClose">
            Cancel
          </Button>
        </DialogClose>
        <Button type="button" @click="handleSubmit">
          {{ props.isUpdate ? "Update" : "Create" }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
