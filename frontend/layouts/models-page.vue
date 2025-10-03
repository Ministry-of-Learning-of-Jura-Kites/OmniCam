<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
import { WORKSPACE_KEY } from "./workspace-provider";

const route = useRoute();

const path = route.fullPath;

// Check if it matches `/workspaces/<id>`
const regex = /workspaces\/([^/]+)\/?$/;
const m = path.match(regex);
const match = m ? m[1] : null;

const workspaceRef = ref(match);

provide(WORKSPACE_KEY, workspaceRef);
</script>

<template>
  <div class="flex flex-col h-screen">
    <SceneStatesProvider
      :project-id="route.params.projectId"
      :model-id="route.params.modelId"
      :workspace="workspaceRef"
    >
      <TopBar :workspace="workspaceRef" />
      <div class="flex-1 flex overflow-hidden">
        <div class="flex-1 w-full h-full">
          <slot />
        </div>
        <CameraPanel :workspace="workspaceRef" />
      </div>
    </SceneStatesProvider>
  </div>
</template>
