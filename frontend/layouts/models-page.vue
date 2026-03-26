<script setup lang="ts">
import LazyTopBar from "@/components/TopBar.vue";
import LazyCameraPanel from "@/components/CameraPanel.vue";
import LazyAlgoPanel from "@/components/AlgoPanel.vue";
import LazyCalibrationPanel from "@/components/CalibrationPanel.vue";
import SceneStatesProvider from "~/components/3d/scene-states-provider/SceneStatesProvider.vue";
import {
  type CamPanelInfo,
  MAP_KEY,
  PANEL_KEY as PANEL_KEY,
  type PanelInfo,
} from "~/constants/state-keys";

import FailDialog from "~/components/dialog/FailDialog.vue";
import { useFailDialog } from "~/composables/useFailDialog";

const { open, message } = useFailDialog();
const route = useRoute();

const isMapOpen = ref(false);

function toggleMap() {
  isMapOpen.value = !isMapOpen.value;
}

provide(MAP_KEY, {
  isMapOpen,
  toggleMap,
});

// Panel Key
const isPanelOpen = ref(true);
const slotWidth = ref("100%");
const currentPanel: PanelInfo["currentPanel"] = ref("camera");
const camPanelSelectedCamId: CamPanelInfo["selectedCamId"] = ref(null);

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
    closePanel();
    return;
  }

  currentPanel.value = "camera";
  openPanel();
}

function toggleAlgoPanel() {
  if (currentPanel.value === "algo" && isPanelOpen.value) {
    closePanel();
    return;
  }

  currentPanel.value = "algo";
  openPanel();
}

function toggleCalibration() {
  isCalibrating.value = !isCalibrating.value;
}

const calibrationGridScale = ref(1);
const isCalibrating = ref(false);

provide(PANEL_KEY, {
  currentPanel,
  togglePanel,
  isPanelOpen,
  toggleAlgoPanel,
  camPanelInfo: { selectedCamId: camPanelSelectedCamId },
  calibrationPanelInfo: {
    isCalibrating,
    toggleCalibration,
    calibrationGridScale,
  },
});

const workspace = computed(() => route.meta.routeInfo?.workspace);
</script>

<template>
  <div class="flex flex-col h-screen">
    <SceneStatesProvider
      :key="`${route.params.projectId}-${route.params.modelId}-${workspace}`"
      :project-id="route.params.projectId as string"
      :model-id="route.params.modelId as string"
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
          <LazyAlgoPanel v-else-if="currentPanel === 'algo'" />
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
