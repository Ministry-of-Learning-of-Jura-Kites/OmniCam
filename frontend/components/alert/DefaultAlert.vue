<script setup lang="ts">
import type { Component } from "vue";
import { X } from "lucide-vue-next";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

defineProps<{
  descriptionMessage: string;
  type: "error" | "success" | "warning";
  icon: Component;
  title?: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();
</script>

<template>
  <div class="fixed top-20 left-1/2 z-50 -translate-x-1/2">
    <Alert
      class="relative w-[420px] shadow-lg pr-10"
      :class="{
        'border-red-500 text-red-500': type === 'error',
        'border-green-500 text-green-500': type === 'success',
        'border-yellow-500 text-yellow-500': type === 'warning',
      }"
    >
      <component
        :is="icon"
        :class="{
          '**:stroke-red-500': type === 'error',
          '**:stroke-green-500': type === 'success',
          '**:stroke-yellow-500': type === 'warning',
        }"
        class="absolute left-4 top-4 h-4 w-4"
      />

      <button
        class="absolute right-3 top-3 opacity-70 transition-opacity hover:opacity-100"
        @click="emit('close')"
      >
        <X class="h-4 w-4" />
      </button>

      <AlertTitle v-if="title">
        {{ title }}
      </AlertTitle>

      <AlertDescription>
        {{ descriptionMessage }}
      </AlertDescription>
    </Alert>
  </div>
</template>
