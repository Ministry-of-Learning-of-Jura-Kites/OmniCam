<template>
  <div :class="classes" v-bind="attrs">
    <slot />
  </div>
</template>

<script setup>
import { useAttrs } from "vue";
import { computed } from "vue";

const attrs = useAttrs();
const props = defineProps({
  class: { type: String, default: "" },
});

const classes = computed(() =>
  [
    "rounded-lg border bg-card text-card-foreground shadow-sm",
    props.class,
  ].join(" "),
);
</script>
