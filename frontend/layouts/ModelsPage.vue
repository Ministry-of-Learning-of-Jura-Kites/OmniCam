<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";

const props = defineProps({
  workspace: {
    type: String,
    required: true,
  },
});

const route = useRoute();
</script>

<template>
  <div class="flex flex-col h-screen">
    <SceneStatesProvider
      :project-id="route.params.projectId"
      :model-id="route.params.modelId"
      :workspace="props.workspace"
    >
      <TopBar :workspace="props.workspace" />
      <div class="flex-1 flex overflow-hidden">
        <div class="flex-1 w-full h-full">
          <slot />
        </div>
        <CameraPanel :workspace="props.workspace" />
      </div>
    </SceneStatesProvider>
  </div>
</template>
