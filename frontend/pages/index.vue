<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { Button } from "@/components/ui/button";
const { data } = await useFetch("/api/test");
</script>
<template>
  <header>
    show all project / include create clone and routing to project id
  </header>
</template>
