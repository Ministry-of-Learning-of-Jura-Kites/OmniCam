<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
</script>

<template>
  <div class="flex flex-col h-screen">
    <TopBar />
    <div class="flex-1 flex">
      <div class="flex-1 w-full h-full">
        <slot />
      </div>
      <CameraPanel />
    </div>
  </div>
</template>
