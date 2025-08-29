<script setup lang="ts" generic="T extends ModelForm">
import { reactive, watch, defineProps, defineEmits } from "vue";
import { Button } from "@/components/ui/button";
import type { FieldConfig } from "./types";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import type { ModelForm } from "~/pages/projects/[projectId]/index.vue";

const props = defineProps<{
  model?: T;
  open: boolean;
  fields: FieldConfig[];
}>();

const emit = defineEmits<{
  (e: "close" | "submit"): void;
  (e: "update:open", value: boolean): void;
}>();

const form = reactive<T>({} as T);

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
          {{ props.model ? "Update" : "Create" }}
        </DialogTitle>
      </DialogHeader>

      <div class="flex flex-col gap-2 py-4">
        <template v-for="field in props.fields" :key="field.key">
          <h2>{{ field.key }}</h2>

          <!-- text/number input -->
          <input
            v-if="field.type === 'text' || field.type === 'number'"
            v-model="(form as any)[field.key]"
            :type="field.type"
            :placeholder="field.key"
            class="border px-2 py-1 rounded"
          />

          <!-- textarea -->
          <textarea
            v-else-if="field.type === 'textarea'"
            v-model="(form as any)[field.key]"
            :placeholder="field.key"
            class="border px-2 py-1 rounded"
          />

          <!-- file input -->
          <input
            v-else-if="field.type === 'file'"
            type="file"
            @change="
              (e) =>
                ((form as any)[field.key] = ((e.target as HTMLInputElement)
                  .files?.[0] ?? null) as any)
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
          {{ props.model ? "Update" : "Create" }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
