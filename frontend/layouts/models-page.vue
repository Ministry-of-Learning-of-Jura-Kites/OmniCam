<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
import { WORKSPACE_KEY } from "./workspace-provider";
console.log("gggg");

const route = useRoute();

const workspace = inject(WORKSPACE_KEY, null);
</script>

<template>
  <div class="flex flex-col h-screen">
    <SceneStatesProvider
      :project-id="route.params.projectId"
      :model-id="route.params.modelId"
      :workspace="workspace"
    >
      <TopBar :workspace="workspace" />
      <div class="flex-1 flex overflow-hidden">
        <div class="flex-1 w-full h-full">
          <NuxtPage />
        </div>
        <CameraPanel :workspace="workspace" />
      </div>
    </SceneStatesProvider>
  </div>
</template>
