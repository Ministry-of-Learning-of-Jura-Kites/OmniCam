<script setup lang="ts" generic="L extends Record<string, InputTypes | null>">
import { watch } from "vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import type { InputTypeMap, InputTypes } from "./types";

type ModelFromFields<L extends Record<string, InputTypes | null>> = {
  [K in keyof L]: L[K] extends InputTypes ? InputTypeMap[L[K]] | null : null;
};

type TitleMap<L extends Record<string, InputTypes | null>> = {
  [K in keyof L]: string;
};

const model = defineModel<ModelFromFields<L>>("model", {
  type: Object,
  default: () =>
    ({
      name: "",
      description: "",
      file: null,
    }) as ModelFromFields<L>,
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

// const model = reactive<ModelFromFields<L>>({} as ModelFromFields<L>);

watch(
  () => props.open,
  (_val) => {
    //if (val && model) {
    //  Object.assign(model, model);
    //}
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
          {{ props.mode === "update" ? "Update" : "Create" }}
        </DialogTitle>
      </DialogHeader>

      <div class="flex flex-col gap-2 py-4">
        <template
          v-for="[key, type] of Object.entries(props.fields) as [
            keyof L,
            L[keyof L],
          ][]"
          :key="key"
        >
          <h2>{{ titles[key] }}</h2>

          <!-- text/number input -->
          <input
            v-if="type === 'text' || type === 'number'"
            v-model="(model as any)[key]"
            :type="type"
            :placeholder="key as string"
            class="border px-2 py-1 rounded"
          />

          <!-- textarea -->
          <textarea
            v-else-if="type === 'textarea'"
            v-model="(model as any)[key]"
            :placeholder="key as string"
            class="border px-2 py-1 rounded"
          />

          <!-- file input -->
          <input
            v-else-if="type === 'file'"
            type="file"
            @change="
              (e) =>
                ((model as any)[key] = ((e.target as HTMLInputElement)
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
          {{ props.mode === "update" ? "Update" : "Create" }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
