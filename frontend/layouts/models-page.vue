<script setup>
import LazyTopBar from "@/components/TopBar.vue";
import LazyCameraPanel from "@/components/CameraPanel.vue";
import LazyAlgoPanel from "@/components/AlgoPanel.vue";
import LazyCalibrationPanel from "@/components/CalibrationPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
import {
  IS_PANEL_OPEN_KEY,
  TOGGLE_PANEL_KEY,
  TOGGLE_MINIMAP_KEY,
  IS_MAP_OPEN_KEY,
  IS_CALIBRATING_KEY,
  TOGGLE_CALIBRATION_KEY,
  CALIBRATION_GRID_SCALE,
  CURRENT_PANEL,
  TOGGLE_ALGO_PANEL_KEY,
} from "~/constants/state-keys";

import FailDialog from "~/components/dialog/FailDialog.vue";
import { useFailDialog } from "~/composables/useFailDialog";

const { open, message } = useFailDialog();
const route = useRoute();

// Panel Key
const isPanelOpen = ref(true);
const slotWidth = ref("100%");
const isMapOpen = ref(false);
const currentPanel = ref("camera");

onMounted(() => {
  slotWidth.value = isPanelOpen.value ? "calc(100% - 20rem)" : "100%";
});

function openPanel() {
  isPanelOpen.value = true;
  slotWidth.value = "calc(100% - 20rem)";
}

function closePanel() {
  isPanelOpen.value = false;
  slotWidth.value = "100%";
}

function togglePanel() {
  if (currentPanel.value === "camera" && isPanelOpen.value) {
    currentPanel.value = null;
    closePanel();
    return;
  }

  currentPanel.value = "camera";
  openPanel();
}

function toggleAlgoPanel() {
  if (currentPanel.value === "algo" && isPanelOpen.value) {
    currentPanel.value = null;
    closePanel();
    return;
  }

  currentPanel.value = "algo";
  openPanel();
}

function toggleMiniMap() {
  isMapOpen.value = !isMapOpen.value;
}

provide(IS_PANEL_OPEN_KEY, isPanelOpen);
provide(TOGGLE_PANEL_KEY, togglePanel);
provide(TOGGLE_ALGO_PANEL_KEY, toggleAlgoPanel);
provide(CURRENT_PANEL, currentPanel);
provide(IS_MAP_OPEN_KEY, isMapOpen);
provide(TOGGLE_MINIMAP_KEY, toggleMiniMap);

// Calibration Key
const isCalibrating = ref(false);
function toggleCalibration() {
  isCalibrating.value = !isCalibrating.value;
}
provide(IS_CALIBRATING_KEY, isCalibrating);
provide(TOGGLE_CALIBRATION_KEY, toggleCalibration);

const calibrationGridScale = ref(1);
provide(CALIBRATION_GRID_SCALE, calibrationGridScale);

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
          <LazyCalibrationPanel v-if="isCalibrating" />
          <LazyCameraPanel
            v-else-if="currentPanel === 'camera'"
            :workspace="workspace"
          />
          <LazyAlgoPanel
            v-else-if="currentPanel === 'algo'"
            :workspace="workspace"
          />
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
