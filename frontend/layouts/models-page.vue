<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";

const route = useRoute();

// const path = route.fullPath;

// Check if it matches `/workspaces/<id>`
// const regex = /workspaces\/([^/]+)\/?$/;
// const m = path.match(regex);
// const match = m ? m[1] : null;

// const workspace = ref(match);

const workspace = computed(() => route.meta.routeInfo?.workspace);

// provide(WORKSPACE_KEY, workspace);
</script>

<template>
  <div class="flex flex-col h-screen">
    <SceneStatesProvider
      :key="`${route.params.projectId}-${route.params.modelId}-${workspace}`"
      :project-id="route.params.projectId"
      :model-id="route.params.modelId"
      :workspace="workspace"
    >
      <TopBar :workspace="workspace" />
      <div class="flex-1 flex overflow-hidden">
        <div class="flex-1 w-full h-full">
          <slot />
        </div>
        <CameraPanel :workspace="workspace" />
      </div>
    </SceneStatesProvider>
  </div>
</template>
