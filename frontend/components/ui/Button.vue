<template>
  <!-- ใช้ tag ตาม props.as (default button) -->
  <component :is="as" :class="classes" v-bind="attrs" v-on="$attrs">
    <slot />
  </component>
</template>

<script setup>
import { computed, useAttrs } from "vue";

const attrs = useAttrs();

const props = defineProps({
  variant: {
    type: String,
    default: "default",
    validator: (v) =>
      [
        "default",
        "destructive",
        "outline",
        "secondary",
        "ghost",
        "link",
      ].includes(v),
  },
  size: {
    type: String,
    default: "default",
    validator: (v) => ["default", "sm", "lg", "icon"].includes(v),
  },
  class: { type: String, default: "" },
  as: { type: String, default: "button" }, // เปลี่ยน tag ถ้าอยาก
});

// CVA classes
const baseClasses =
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0";

const variantClasses = {
  default: "bg-primary text-primary-foreground hover:bg-primary/90",
  destructive:
    "bg-destructive text-destructive-foreground hover:bg-destructive/90",
  outline:
    "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
  secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
  ghost: "hover:bg-accent hover:text-accent-foreground",
  link: "text-primary underline-offset-4 hover:underline",
};

const sizeClasses = {
  default: "h-10 px-4 py-2",
  sm: "h-9 rounded-md px-3",
  lg: "h-11 rounded-md px-8",
  icon: "h-10 w-10",
};

const classes = computed(() =>
  [
    baseClasses,
    variantClasses[props.variant],
    sizeClasses[props.size],
    props.class,
  ].join(" "),
);
</script>
