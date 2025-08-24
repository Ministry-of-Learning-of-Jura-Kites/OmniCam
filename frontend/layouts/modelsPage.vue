<template>
  <div class="flex flex-col h-screen">
    <TopBar />
    <div class="flex-1 flex">
      <div class="flex-1">
        <slot />
      </div>
      <!-- <CameraPanel /> -->
    </div>
  </div>
</template>

<script setup>
import TopBar from "@/components/TopBar.vue";
//import CameraPanel from "@/components/CameraPanel.vue";
</script>
