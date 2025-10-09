<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";

const route = useRoute();

const isPanelOpen = ref(true);
const slotWidth = ref("100%");

onMounted(() => {
  slotWidth.value = isPanelOpen.value ? "calc(100% - 20rem)" : "100%";
});

function togglePanel() {
  isPanelOpen.value = !isPanelOpen.value;
  slotWidth.value = isPanelOpen.value ? "calc(100% - 20rem)" : "100%";
}

provide("isPanelOpen", isPanelOpen);
provide("togglePanel", togglePanel);

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
        <div
          class="h-full transition-all duration-300"
          :style="{ width: slotWidth }"
        >
          <slot />
        </div>

        <div
          class="h-full transition-all duration-300 overflow-hidden"
          :style="{ width: isPanelOpen ? '20rem' : '0' }"
        >
          <CameraPanel />
        </div>
      </div>
    </SceneStatesProvider>
  </div>
</template>
