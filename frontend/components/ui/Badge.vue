<template>
  <div :class="classes" v-bind="attrs">
    <slot />
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useAttrs } from "vue";

// ใช้ attrs เพื่อรองรับ HTML attributes เช่น id, title, style
const attrs = useAttrs();

// Props
const props = defineProps({
  variant: {
    type: String,
    default: "default",
    validator: (v) =>
      ["default", "secondary", "destructive", "outline"].includes(v),
  },
  class: {
    type: String,
    default: "",
  },
});

// Function คล้าย cva ของ React
const baseClasses =
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2";

const variantClasses = {
  default:
    "border-transparent bg-primary text-primary-foreground hover:bg-primary/80",
  secondary:
    "border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",
  destructive:
    "border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80",
  outline: "text-foreground",
};

const classes = computed(() =>
  [baseClasses, variantClasses[props.variant], props.class].join(" "),
);
</script>
