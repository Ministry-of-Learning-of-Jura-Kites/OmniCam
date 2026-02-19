<script setup lang="ts">
import LazyTopBar from "@/components/TopBar.vue";
import LazyCameraPanel from "@/components/CameraPanel.vue";
import LazyAlgoPanel from "@/components/AlgoPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
import {
  TOGGLE_CAM_PANEL_KEY,
  IS_MAP_OPEN_KEY,
  TOGGLE_MINIMAP_KEY,
  CURRENT_PANEL,
  TOGGLE_ALGO_PANEL_KEY,
  WORKSPACE,
} from "~/constants/state-keys";

import FailDialog from "~/components/dialog/FailDialog.vue";
import { useFailDialog } from "~/composables/useFailDialog";
const { open, message } = useFailDialog();

const route = useRoute();

const currentPanel = ref<"camera" | "algo" | null>("camera");
const isMapOpen = ref(false);

const slotWidth = computed(() => {
  return currentPanel.value != null ? "calc(100% - 20rem)" : "100%";
});

const panelsWidth = computed(() => {
  return {
    camera: currentPanel.value == "camera" ? "20rem" : "0",
    algo: currentPanel.value == "algo" ? "20rem" : "0",
  };
});

function toggleCamPanel() {
  if (currentPanel.value == "camera") {
    currentPanel.value = null;
  } else {
    currentPanel.value = "camera";
  }
}

function toggleMiniMap() {
  isMapOpen.value = !isMapOpen.value;
}

function toggleAlgoPanel() {
  if (currentPanel.value == "algo") {
    currentPanel.value = null;
  } else {
    currentPanel.value = "algo";
  }
}

provide(CURRENT_PANEL, currentPanel);
provide(TOGGLE_CAM_PANEL_KEY, toggleCamPanel);
provide(TOGGLE_ALGO_PANEL_KEY, toggleAlgoPanel);

provide(IS_MAP_OPEN_KEY, isMapOpen);
provide(TOGGLE_MINIMAP_KEY, toggleMiniMap);

const workspace = route.meta.routeInfo?.workspace ?? null;
provide(WORKSPACE, workspace);
</script>

<template>
  <div class="flex flex-col h-screen">
    <SceneStatesProvider
      :key="`${route.params.projectId}-${route.params.modelId}-${workspace}`"
      :project-id="route.params.projectId as string"
      :model-id="route.params.modelId as string"
    >
      <LazyTopBar />
      <div class="flex-1 flex overflow-hidden">
        <div
          class="h-full transition-all duration-300"
          :style="{ width: slotWidth }"
        >
          <slot />
        </div>

        <div
          class="h-full transition-all duration-300 overflow-hidden"
          :style="{ width: panelsWidth.camera }"
        >
          <LazyCameraPanel />
        </div>

        <div
          class="h-full transition-all duration-300 overflow-hidden"
          :style="{ width: panelsWidth.algo }"
        >
          <LazyAlgoPanel />
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
