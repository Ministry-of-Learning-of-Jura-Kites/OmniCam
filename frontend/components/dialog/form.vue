<script setup lang="ts">
import { reactive, watch, defineProps, defineEmits } from "vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

type InputTypes = "text" | "number" | "textarea" | "file";

export interface FieldConfig {
  key: string;
  type: InputTypes;
}

const props = defineProps<{
  model?: Record<string, any>;
  open: boolean;
  fields: FieldConfig[];
  isUpdate?: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "submit"): void;
  (e: "update:open", value: boolean): void;
}>();

const form = reactive<Record<string, any>>({});

// populate form if editing
watch(
  () => props.open,

  (val) => {
    console.log(props.open, "check");
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
          <input
            v-if="field.type === 'text' || field.type === 'number'"
            v-model="form[field.key]"
            :type="field.type"
            :placeholder="field.key"
            class="border px-2 py-1 rounded"
          />
          <textarea
            v-else-if="field.type === 'textarea'"
            v-model="form[field.key]"
            :placeholder="field.key"
            class="border px-2 py-1 rounded"
          />
          <input
            v-else-if="field.type === 'file'"
            type="file"
            @change="
              (e) =>
                (form[field.key] =
                  (e.target as HTMLInputElement).files?.[0] ?? null)
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
