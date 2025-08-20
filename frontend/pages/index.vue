<script setup lang="ts">
const { data } = await useFetch("/api/test");
</script>
<template>
  <span class="bg-green-100">test {{ data }}</span>
</template>
