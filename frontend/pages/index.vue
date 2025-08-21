<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { useFetch } from "nuxt/app";

// import { useFetch } from "nuxt/app";

const { data } = await useFetch("/api/test");
</script>
<template>
  <span class="bg-green-100">test {{ data }}</span>
</template>
