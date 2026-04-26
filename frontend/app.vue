<script setup lang="ts">
const { theme } = useLightDarkTheme();

const definedTheme = computed(() => {
  return theme.value ?? "light";
});

useHead({
  htmlAttrs: {
    class: definedTheme, // or dynamic variable
  },
});
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
