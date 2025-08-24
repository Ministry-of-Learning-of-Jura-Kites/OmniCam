<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { Button } from "@/components/ui/button";
const { data } = await useFetch("/api/test");
</script>
<template>
  <div>
    <span class="bg-green-100">test {{ data }}</span>
    <Button>Button</Button>
  </div>
</template>
