<script setup>
import TopBar from "@/components/TopBar.vue";
import CameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
import { IS_PANEL_OPEN_KEY, TOGGLE_PANEL_KEY } from "~/constants/state-keys";

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

provide(IS_PANEL_OPEN_KEY, isPanelOpen);
provide(TOGGLE_PANEL_KEY, togglePanel);

const workspace = computed(() => route.meta.routeInfo?.workspace);
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
          <CameraPanel :workspace="workspace" />
        </div>
      </div>
    </SceneStatesProvider>
  </div>
</template>
