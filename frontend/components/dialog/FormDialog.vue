<script setup lang="ts" generic="L extends Record<string, FieldOption | null>">
import { ref } from "vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import type { InputTypeMap, InputTypes, FieldOption } from "./types";

type ModelFromFields<L extends Record<string, FieldOption | null>> = {
  [K in keyof L]: L[K] extends { type: InputTypes }
    ? InputTypeMap[L[K]["type"]] | null
    : null;
};

type TitleMap<L extends Record<string, FieldOption | null>> = {
  [K in keyof L]: string;
};

const model = defineModel<ModelFromFields<L>>("model", {
  type: Object,
  default: () => ({}) as ModelFromFields<L>,
});

const props = defineProps<{
  open: boolean;
  titles: TitleMap<L>;
  fields: L;
  mode: "create" | "update";
}>();

const emit = defineEmits<{
  (e: "close" | "submit"): void;
  (e: "update:open", value: boolean): void;
}>();

const errors = ref<Record<string, string>>({});

function validate(): boolean {
  const newErrors: Record<string, string> = {};

  for (const [key, field] of Object.entries(props.fields)) {
    if (!field) continue;
    if (field.required && !model.value[key as keyof typeof model.value]) {
      newErrors[key] = `${props.titles[key as keyof L]} is required.`;
    }
  }

  errors.value = newErrors;
  return Object.keys(newErrors).length === 0;
}

function handleSubmit() {
  if (!validate()) return;
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
          {{ props.mode === "update" ? "Update" : "Create" }}
        </DialogTitle>
      </DialogHeader>

      <div class="flex flex-col gap-4 py-4">
        <template
          v-for="[key, field] of Object.entries(props.fields) as [
            keyof L,
            L[keyof L],
          ][]"
          :key="key"
        >
          <h2>
            {{ props.titles[key] }}
            <span v-if="field?.required" class="text-red-500"> * </span>
          </h2>
          <p v-if="errors[key as string]" class="text-sm text-red-500">
            {{ errors[key as string] }}
          </p>

          <!-- text/number input -->
          <input
            v-if="field?.type === 'text' || field?.type === 'number'"
            v-model="(model as any)[key]"
            :type="field?.type"
            :placeholder="key as string"
            :class="[
              'border px-2 py-1 rounded',
              errors[key as string] ? 'border-red-500' : 'border-gray-300',
            ]"
          />

          <textarea
            v-else-if="field?.type === 'textarea'"
            v-model="(model as any)[key]"
            :placeholder="key as string"
            :class="[
              'border px-2 py-1 rounded',
              errors[key as string] ? 'border-red-500' : 'border-gray-300',
            ]"
          />

          <input
            v-else-if="field?.type === 'file'"
            type="file"
            :accept="
              key === 'file'
                ? '.glb'
                : key === 'image'
                  ? '.jpg,.png'
                  : undefined
            "
            @change="
              (e) =>
                ((model as any)[key] = ((e.target as HTMLInputElement)
                  .files?.[0] ?? null) as any)
            "
            :class="[
              'border px-2 py-1 rounded',
              errors[key as string] ? 'border-red-500' : 'border-white',
            ]"
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
          {{ props.mode === "update" ? "Update" : "Create" }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
