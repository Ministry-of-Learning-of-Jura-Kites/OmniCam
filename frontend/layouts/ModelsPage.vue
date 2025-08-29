<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
</script>

<template>
  <div class="flex flex-col h-screen">
    <TopBar />
    <SceneStatesProvider>
      <div class="flex-1 flex">
        <div class="flex-1 w-full h-full">
          <slot />
        </div>
        <CameraPanel />
      </div>
    </SceneStatesProvider>
  </div>
</template>
