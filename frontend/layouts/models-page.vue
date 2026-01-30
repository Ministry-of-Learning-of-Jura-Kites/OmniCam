<script setup>
import LazyTopBar from "@/components/TopBar.vue";
import LazyCameraPanel from "@/components/CameraPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
import {
  IS_PANEL_OPEN_KEY,
  TOGGLE_PANEL_KEY,
  IS_MAP_OPEN_KEY,
  TOGGLE_MINIMAP_KEY,
} from "~/constants/state-keys";

import FailDialog from "~/components/dialog/FailDialog.vue";
import { useFailDialog } from "~/composables/useFailDialog";
const { open, message } = useFailDialog();

const route = useRoute();

const isPanelOpen = ref(true);
const slotWidth = ref("100%");
const isMapOpen = ref(false);

onMounted(() => {
  slotWidth.value = isPanelOpen.value ? "calc(100% - 20rem)" : "100%";
});

function togglePanel() {
  isPanelOpen.value = !isPanelOpen.value;
  slotWidth.value = isPanelOpen.value ? "calc(100% - 20rem)" : "100%";
}

function toggleMiniMap() {
  isMapOpen.value = !isMapOpen.value;
}

provide(IS_PANEL_OPEN_KEY, isPanelOpen);
provide(TOGGLE_PANEL_KEY, togglePanel);
provide(IS_MAP_OPEN_KEY, isMapOpen);
provide(TOGGLE_MINIMAP_KEY, toggleMiniMap);

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
      <LazyTopBar :workspace="workspace" />
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
          <LazyCameraPanel :workspace="workspace" />
        </div>
      </div>

      <FailDialog
        v-model:open="open"
        :message="message"
        icon="fa-solid fa-circle-exclamation"
      />
    </SceneStatesProvider>
  </div>
</template>
