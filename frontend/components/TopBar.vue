<script setup lang="ts">
import Button from "./ui/button/Button.vue";
import Badge from "./ui/badge/Badge.vue";
import Card from "./ui/card/Card.vue";

import {
  PackageOpen,
  RefreshCcw,
  Maximize,
  CloudCheck,
  Download,
  Upload,
  Save,
  User,
  LogOut,
  IndentDecrease,
  IndentIncrease,
  EllipsisVertical,
  Trash2,
  Moon,
  Sun,
  Map,
  Settings2,
  RulerDimensionLine,
  Check,
  X,
} from "lucide-vue-next";

import { exportCamerasToJson } from "@/utils/exportScene";
import { importJsonToCameras } from "@/utils/importScene";
import { PANEL_KEY, SCENE_STATES_KEY } from "@/constants/state-keys";

import Tooltip from "./ui/tooltip/Tooltip.vue";
import TooltipTrigger from "./ui/tooltip/TooltipTrigger.vue";
import TooltipContent from "./ui/tooltip/TooltipContent.vue";
import TooltipProvider from "./ui/tooltip/TooltipProvider.vue";
import { MODEL_INFO_KEY, MAP_KEY } from "~/constants/state-keys";
import Setting3dDialog from "./dialog/Setting3dDialog.vue";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const candidateEntries = computed(() =>
  Object.entries(sceneStates.optimization?.candidateCameras ?? {}),
);

const {
  currentPanel,
  toggleAlgoPanel,
  togglePanel,
  calibrationPanelInfo,
  isPanelOpen,
} = inject(PANEL_KEY)!;

const { toggleCalibration, isCalibrating } = calibrationPanelInfo;

const sceneStates = inject(SCENE_STATES_KEY)!;

const { isMapOpen, toggleMap } = inject(MAP_KEY)!;

const route = useRoute();

const openDialog = ref(false);

const dialogTitle = ref("");
const dialogContent = ref("");

const lightDarkTheme = useLightDarkTheme();

const openResolver = ref(false);
const conflicts = ref({});

const isSettingDialogOpen = ref<boolean>(false);

const isCameraActive = computed(() => sceneStates.currentCamId.value !== null);

async function saveModelToPublic() {
  // Mocked data
  // const respJson = {
  //   noChanges: false,
  //   conflicts: {
  //     "c6a9ff22-0962-4d88-bbbd-2053d883f43b": {
  //       angleX: {
  //         base: 1,
  //         main: 2,
  //         workspace: 1,
  //       },
  //       angleY: {
  //         base: 1,
  //         main: 2,
  //         workspace: 1,
  //       },
  //       angleZ: {
  //         base: 1,
  //         main: 2,
  //         workspace: 1,
  //       },
  //       test: {
  //         base: {
  //           "123": 456,
  //         },
  //         main: {
  //           "123": 456,
  //         },
  //         workspace: {
  //           "123": 451,
  //         },
  //       },
  //     },
  //   },
  // };

  const runtimeConfig = useRuntimeConfig();
  const resp = await fetch(
    `http://${runtimeConfig.public.externalBackendHost}/api/v1/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me/merge`,
    { method: "POST", credentials: "include" },
  );

  if (!resp.ok) {
    console.error(resp);
    return;
  }

  const respJson: {
    noChanges?: boolean;
    conflicts: Record<string, Record<string, unknown>>;
  } = await resp.json();

  if (respJson.noChanges) {
    dialogTitle.value = "No changes";
    dialogContent.value = "There is no changes to be published";
    openDialog.value = true;
    return;
  }
  if (respJson.conflicts) {
    conflicts.value = respJson.conflicts;
    openResolver.value = true;
  } else {
    dialogTitle.value = "Progress Saved!";
    dialogContent.value = "";
    openDialog.value = true;
  }
}

function openFileDialog() {
  const input = document.createElement("input");
  input.type = "file";
  input.accept = ".json";
  input.style.display = "none";

  input.addEventListener("change", async (event: Event) => {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];
    if (!file) return;

    try {
      const text = await file.text();
      importJsonToCameras(sceneStates, text);
    } catch (err) {
      console.error("Failed to import cameras:", err);
    }
  });

  document.body.appendChild(input);
  input.click();
  input.remove();
}

function toggleSettingDialog() {
  isSettingDialogOpen.value = true;
}

async function createWorkspace() {
  const runtimeConfig = useRuntimeConfig();

  try {
    const data = await $fetch(
      `http://${runtimeConfig.public.externalBackendHost}/api/v1/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me`,
      {
        method: "POST",
        credentials: "include",
      },
    );
    useState(`${MODEL_INFO_KEY}-${route.params.modelId}`, () => data);
    goToMyWorkspace();
  } catch (err) {
    console.error(err);
    showError({
      message: "Failed to create workspace",
    });
  }
}

const runtimeConfig = useRuntimeConfig();

async function deleteWorkspace() {
  try {
    await $fetch<undefined>(
      `http://${runtimeConfig.public.externalBackendHost}/api/v1/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me`,
      {
        credentials: "include",
        method: "DELETE",
      },
    );

    useState(`${MODEL_INFO_KEY}-${route.params.modelId}`, () => undefined);
    goToModel();
  } catch (err) {
    console.error(err);
    showError({
      message: "Failed to create workspace",
    });
  }
}

function goToModel() {
  navigateTo(
    `/projects/${route.params.projectId}/models/${route.params.modelId}`,
  );
}
function goToMyWorkspace() {
  navigateTo(
    `/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me`,
  );
}

function confirmOptimizedCams() {
  for (const [camId, candidate] of candidateEntries.value) {
    sceneStates.cameras[camId] = candidate;
  }
  removeOptimizedCams();
}
function removeOptimizedCams() {
  for (const [camId, _candidate] of candidateEntries.value) {
    delete sceneStates.optimization!.candidateCameras[camId];
  }
}
</script>

<template>
  <MergeConflictsResolver
    :visible="openResolver"
    :conflicts="conflicts"
    @resolved="goToModel()"
    @close="openResolver = false"
  />

  <Dialog :open="openDialog" @update:open="openDialog = $event">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>{{ dialogTitle }}</DialogTitle>
        <DialogDescription> {{ dialogContent }} </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button type="submit" @click="openDialog = !openDialog">Ok</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <TooltipProvider>
    <div
      class="h-16 border-b border-border px-6 flex items-center justify-between"
    >
      <!-- Project Info -->
      <div id="left-menu" class="flex items-center gap-4">
        <Card class="px-3 py-1 bg-survey-surface">
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 bg-survey-accent rounded-full" />
            <span class="text-sm font-medium max-w-[50px] truncate">{{
              sceneStates.modelInfo.data.name
            }}</span>
            <Badge variant="secondary" class="ml-2">
              {{ props.workspace == null ? "Public" : "Workspace" }}
            </Badge>
          </div>
        </Card>
        <div class="flex items-center justify-center">
          <Tooltip
            v-if="
              workspace == 'me' &&
              (sceneStates.markedForCheck.value ||
                sceneStates.localVersion.value !==
                  sceneStates.lastSyncedVersion.value ||
                sceneStates.calibration.dirty ||
                sceneStates.markedFacesForCheck.value)
            "
          >
            <TooltipTrigger>
              <RefreshCcw
                class="animate-spin"
                style="animation-direction: reverse"
            /></TooltipTrigger>
            <TooltipContent> Saving </TooltipContent>
          </Tooltip>
          <Tooltip v-else>
            <TooltipTrigger><CloudCheck /></TooltipTrigger>
            <TooltipContent> Saved to Cloud </TooltipContent>
          </Tooltip>
        </div>
      </div>

      <!-- Scene Controls -->
      <div id="middle-menu" class="flex items-center gap-2">
        <Button
          size="sm"
          variant="outline"
          :disabled="!isCameraActive"
          :class="isCameraActive ? 'bg-red-500! hover:bg-red-700!' : ''"
          @click="sceneStates.cameraManagement.switchToSpectator()"
          ><LogOut class="button-icon" />
          <span class="ml-2 button-span-text">Exit Camera</span>
        </Button>

        <Button size="sm" variant="outline">
          <ClientOnly>
            <Tooltip>
              <TooltipTrigger> <Maximize class="button-icon" /></TooltipTrigger>
              <TooltipContent> Fullscreen </TooltipContent>
            </Tooltip>
          </ClientOnly>
        </Button>

        <div class="h-6 w-px bg-border mx-2" />

        <Button
          v-if="workspace == 'me'"
          size="sm"
          variant="outline"
          @click="saveModelToPublic()"
        >
          <Save class="button-icon" />
          <span class="ml-2 button-span-text"> Publish </span>
        </Button>

        <Button
          v-if="workspace == 'me'"
          size="sm"
          variant="outline"
          @click="
            goToModel();
            isCalibrating = false;
          "
        >
          <LogOut class="button-icon" />
          <span class="button-span-text"> Exit Workspace </span>
        </Button>

        <template v-if="workspace == null">
          <Button
            v-if="sceneStates.modelInfo.data.workspaceExists"
            size="sm"
            variant="outline"
            @click="goToMyWorkspace()"
          >
            <PackageOpen class="button-icon" />
            <span class="ml-2 button-span-text"> Open Workspace </span>
          </Button>
          <Button v-else size="sm" variant="outline" @click="createWorkspace()">
            <PackageOpen class="button-icon" />
            <span class="ml-2 button-span-text"> Create Workspace </span>
          </Button>
        </template>

        <Button
          v-if="workspace == 'me'"
          size="sm"
          variant="outline"
          :class="{ 'btn-calibrating': isCalibrating }"
          @click="() => toggleCalibration()"
        >
          <RulerDimensionLine class="button-icon" />
          <span v-if="!isCalibrating" class="ml-2 button-span-text">
            Calibration</span
          >
          <span v-else class="ml-2 button-span-text"> Calibrating...</span>
        </Button>
        <Button
          size="sm"
          variant="outline"
          :class="isMapOpen ? 'bg-green-400! hover:bg-green-500!' : ''"
          @click="() => toggleMap()"
        >
          <Map class="h-4 w-4" />
          Map
        </Button>
        <Button
          v-if="workspace == 'me'"
          size="sm"
          variant="outline"
          @click="() => toggleAlgoPanel()"
        >
          <IndentIncrease v-if="currentPanel == 'algo'" class="button-icon" />
          <IndentDecrease v-else class="button-icon" />
          <span class="ml-2 button-span-text"> Algo </span>
        </Button>

        <Button
          v-if="candidateEntries.length != 0"
          @click="confirmOptimizedCams"
        >
          <Check /> Confirm
        </Button>
        <Button
          v-if="candidateEntries.length != 0"
          @click="removeOptimizedCams"
        >
          <X /> Cancel
        </Button>
      </div>

      <div
        id="right-menu"
        class="flex flex-row justify-center items-center gap-2"
      >
        <Button size="sm" variant="outline" @click="() => togglePanel()">
          <IndentIncrease
            v-if="isPanelOpen && currentPanel === 'camera'"
            class="button-icon"
          />
          <IndentDecrease v-else class="button-icon" />
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button size="sm" variant="outline">
              <EllipsisVertical class="button-icon" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <!-- <DropdownMenuLabel>{{ user?.first_name }}</DropdownMenuLabel> -->
            <DropdownMenuGroup>
              <DropdownMenuItem @click="() => openFileDialog()">
                <Download class="button-icon" />
                Import
              </DropdownMenuItem>
              <DropdownMenuItem
                @click="() => exportCamerasToJson(sceneStates.cameras)"
              >
                <Upload class="button-icon" />
                Export
              </DropdownMenuItem>
              <DropdownMenuItem @click="toggleSettingDialog">
                <Settings2 class="button-icon" />
                Setting
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <!-- <DropdownMenuSeparator /> -->
            <DropdownMenuGroup>
              <DropdownMenuItem
                v-if="workspace == 'me'"
                @click="deleteWorkspace()"
              >
                <Trash2 class="button-icon" />
                <span class="button-span-text"> Delete Workspace </span>
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>

        <Setting3dDialog
          :open="isSettingDialogOpen"
          message="Update Sensitivity"
          @update:open="isSettingDialogOpen = $event"
        />

        <Button size="sm" variant="outline" @click="lightDarkTheme.toggleTheme">
          <Moon
            v-if="lightDarkTheme.theme.value == 'light'"
            class="button-icon"
          />
          <Sun v-else class="button-icon" />
        </Button>

        <!-- User Actions -->
        <div class="flex items-center gap-2">
          <Button size="sm" variant="outline">
            <User class="button-icon" />
          </Button>
        </div>
      </div>
    </div>
  </TooltipProvider>
</template>

<style lang="scss" scoped>
.button-icon {
  width: calc(4 * 0.25rem);
  height: calc(4 * 0.25rem);
}
.button-span-text {
  display: none;
  text-overflow: ellipsis;
  @media (width >= 53rem) {
    display: inline;
  }
}
Button {
  transition-property:
    color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}
.btn-calibrating {
  background-color: #ef4444 !important; /* สีแดง (Tailwind red-500) */
  color: white !important;
  border-color: #dc2626 !important;
}
.btn-calibrating:hover {
  background-color: #dc2626 !important;
  opacity: 0.9;
}
</style>
