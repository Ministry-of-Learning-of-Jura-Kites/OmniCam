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
  SCENE_STATES_READY_KEY,
  SCENE_STATES_KEY,
} from "~/constants/state-keys";
import FailDialog from "~/components/dialog/FailDialog.vue";
import { useFailDialog } from "~/composables/useFailDialog";
import type { SceneStatesWithHelper } from "~/types/scene-states";

const sceneStatesReady = ref(false);
provide(SCENE_STATES_READY_KEY, sceneStatesReady);

const sceneStates = shallowRef<SceneStatesWithHelper | undefined>(undefined);
provide(SCENE_STATES_KEY, sceneStates);

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
const currentToolMode = ref<"orbit" | "measurement" | "calibration">("orbit");

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
  if (isPanelOpen.value) {
    closePanel();
  } else {
    openPanel();
  }
}

function toggleCameraPanel() {
  currentPanel.value = "camera";
  currentToolMode.value = "orbit";
  openPanel();
}

function toggleAlgoPanel() {
  currentToolMode.value = "orbit";

  if (currentPanel.value === "algo" && isPanelOpen.value) {
    currentPanel.value = "camera";
  } else {
    currentPanel.value = "algo";
    openPanel();
  }
}

function toggleCalibration() {
  if (currentPanel.value === "calibration" && isPanelOpen.value) {
    currentPanel.value = "camera";
    currentToolMode.value = "orbit";
  } else {
    currentPanel.value = "calibration";
    currentToolMode.value = "calibration";
    openPanel();
  }
}

function toggleMeasurement() {
  if (currentPanel.value === "measurement" && isPanelOpen.value) {
    currentPanel.value = "camera";
    currentToolMode.value = "orbit";
  } else {
    currentPanel.value = "measurement";
    currentToolMode.value = "measurement";
    openPanel();
  }
}
const calibrationGridScale = ref(1);

provide(PANEL_KEY, {
  currentPanel,
  currentToolMode,
  togglePanel,
  isPanelOpen,
  toggleAlgoPanel,
  toggleCameraPanel,
  camPanelInfo: { selectedCamId: camPanelSelectedCamId },
  calibrationPanelInfo: {
    toggleCalibration,
    calibrationGridScale,
  },
  toggleMeasurement,
});

const workspace = computed(() => route.params.workspaceId as string);

const isInCameraView = computed(() => {
  return sceneStates?.value?.currentCamId.value !== null;
});

const showPanelWarning = computed(() => {
  return isInCameraView.value && currentPanel.value !== "camera";
});
</script>

<template>
  <div class="flex flex-col h-screen">
    <SceneStatesProvider
      :key="`${route.params.projectId}-${route.params.modelId}-${workspace}`"
      :project-id="route.params.projectId as string"
      :model-id="route.params.modelId as string"
      :workspace="workspace!"
    >
      <LazyTopBar :workspace="workspace!" />

      <div class="flex-1 flex overflow-hidden">
        <div
          v-if="sceneStatesReady"
          class="h-full transition-all duration-300"
          :style="{ width: slotWidth }"
        >
          <slot />
        </div>

        <div
          v-if="sceneStatesReady"
          class="h-full transition-all duration-300 overflow-hidden"
          :style="{ width: isPanelOpen ? '20rem' : '0' }"
        >
          <div
            v-if="showPanelWarning"
            class="w-full bg-red-500/10 border-b border-red-500/20 px-4 py-2 flex items-center justify-between group"
          >
            <div class="flex items-center gap-2">
              <div class="relative flex h-2 w-2">
                <span
                  class="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"
                ></span>
                <span
                  class="relative inline-flex rounded-full h-2 w-2 bg-red-600"
                ></span>
              </div>

              <span
                class="text-[10px] font-bold uppercase tracking-[0.2em] text-red-600 dark:text-red-500"
              >
                Active Camera View
              </span>
            </div>

            <button
              class="text-[9px] font-black uppercase tracking-widest text-red-600 dark:text-red-500 hover:text-red-700 underline underline-offset-4 decoration-red-500/30 hover:decoration-red-500 transition-all"
              @click="toggleCameraPanel"
            >
              Return to Camera
            </button>
          </div>

          <LazyCalibrationPanel v-if="currentPanel == 'calibration'" />
          <LazyCameraPanel
            v-else-if="currentPanel === 'camera'"
            :workspace="workspace"
          />
          <LazyAlgoPanel v-else-if="currentPanel === 'algo'" />
          <LazyMeasurementPanel v-else-if="currentPanel === 'measurement'" />
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
