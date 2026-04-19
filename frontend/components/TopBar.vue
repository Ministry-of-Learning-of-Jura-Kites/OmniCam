<script setup lang="ts">
import Button from "./ui/button/Button.vue";
import Badge from "./ui/badge/Badge.vue";
import Card from "./ui/card/Card.vue";
import {
  PackageOpen,
  ArrowLeft,
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
  Share2,
  Copy,
  HelpCircle,
  Cpu,
  Camera,
} from "lucide-vue-next";

import { exportCamerasToJson } from "@/utils/exportScene";
import { importJsonToCameras } from "@/utils/importScene";
import { PANEL_KEY, SCENE_STATES_KEY } from "@/constants/state-keys";

import Tooltip from "./ui/tooltip/Tooltip.vue";
import TooltipTrigger from "./ui/tooltip/TooltipTrigger.vue";
import TooltipContent from "./ui/tooltip/TooltipContent.vue";
import TooltipProvider from "./ui/tooltip/TooltipProvider.vue";
import { MAP_KEY } from "~/constants/state-keys";
import Setting3dDialog from "./dialog/Setting3dDialog.vue";
import { uuidToBase64Url } from "~/lib/uuid";
import { useWorkspaceApi } from "~/composables/api/use-workspace-api";
import { useAuth } from "~/composables/api/use-auth";
import { useClipboard } from "@vueuse/core";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const route = useRoute();
const runtimeConfig = useRuntimeConfig();

const projectId = route.params.projectId as string;
const modelId = route.params.modelId as string;

const { postCreateMe, postMerge, deleteWorkspaceMe } = useWorkspaceApi(
  projectId,
  modelId,
  runtimeConfig,
);

const candidateEntries = computed(() =>
  Object.entries(sceneStates?.value!.optimization?.candidateCameras ?? {}),
);

const {
  currentPanel,
  toggleCameraPanel,
  toggleAlgoPanel,
  togglePanel,
  calibrationPanelInfo,
  isPanelOpen,
} = inject(PANEL_KEY)!;

const { toggleCalibration } = calibrationPanelInfo;
// Scenestates can be null to allow skeleton UI
const sceneStates = inject(SCENE_STATES_KEY);

const { isMapOpen, toggleMap } = inject(MAP_KEY)!;

const isShortcutsDialogOpen = ref(false);

function toggleShortcutsDialog() {
  isShortcutsDialogOpen.value = true;
}

const openDialog = ref(false);
const isShareDialogOpen = ref(false);
const { copy, copied } = useClipboard();

const dialogTitle = ref("");
const dialogContent = ref("");

const lightDarkTheme = useLightDarkTheme();

const openResolver = ref(false);
const conflicts = ref({});

const isSettingDialogOpen = ref<boolean>(false);

const isCameraActive = ref(false);

const { user, fetchUser } = useAuth();
fetchUser();

onMounted(() => {
  watch(
    () => sceneStates?.value?.currentCamId?.value,
    (currentCamId) => {
      isCameraActive.value = currentCamId !== null;
    },
  );
});

async function saveModelToPublic() {
  const { resp } = await postMerge();
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
      importJsonToCameras(sceneStates!.value!, text);
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
  try {
    await postCreateMe();
    goToMyWorkspace();
  } catch (err) {
    console.error(err);
    showError({
      message: "Failed to create workspace",
    });
  }
}

async function deleteWorkspace() {
  try {
    deleteWorkspaceMe();
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
function goToModels() {
  navigateTo(`/projects/${route.params.projectId}`);
}
function goToMyWorkspace() {
  navigateTo(
    `/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me`,
  );
}

function confirmOptimizedCams() {
  for (const [camId, candidate] of candidateEntries.value) {
    sceneStates!.value!.cameras[camId] = candidate;
  }
  removeOptimizedCams();
}
function removeOptimizedCams() {
  for (const [camId, _candidate] of candidateEntries.value) {
    delete sceneStates!.value!.optimization!.candidateCameras[camId];
  }
}

const shareUrl = computed(() => {
  const baseUrl = window.location.origin;
  const projectId = route.params.projectId as string;
  const modelId = route.params.modelId;
  const userIdBase64 = uuidToBase64Url(user.value!.id);
  return `${baseUrl}/projects/${projectId}/models/${modelId}/workspaces/${userIdBase64}`;
});

async function openShareDialog(open: boolean) {
  isShareDialogOpen.value = open;
  if (open) {
    await copy(shareUrl.value);
  }
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    // If no element is in fullscreen, request it
    document.documentElement.requestFullscreen().catch((err) => {
      console.error(`Error attempting to enable fullscreen: ${err.message}`);
    });
  } else {
    // If already in fullscreen, exit it
    document.exitFullscreen();
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

  <Dialog
    :open="isShortcutsDialogOpen"
    @update:open="isShortcutsDialogOpen = $event"
  >
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>Keyboard Shortcuts</DialogTitle>
        <DialogDescription>
          Use these keys to navigate the 3D scene.
        </DialogDescription>
      </DialogHeader>

      <div class="grid gap-4 py-4">
        <div class="grid grid-cols-2 items-center gap-4">
          <div class="flex gap-1">
            <kbd class="px-2 py-1 bg-muted rounded border text-xs font-mono"
              >W</kbd
            >
            <kbd class="px-2 py-1 bg-muted rounded border text-xs font-mono"
              >A</kbd
            >
            <kbd class="px-2 py-1 bg-muted rounded border text-xs font-mono"
              >S</kbd
            >
            <kbd class="px-2 py-1 bg-muted rounded border text-xs font-mono"
              >D</kbd
            >
          </div>
          <span class="text-sm">Move Camera</span>
        </div>

        <div class="grid grid-cols-2 items-center gap-4">
          <div class="flex gap-1">
            <kbd class="px-2 py-1 bg-muted rounded border text-xs font-mono"
              >Shift</kbd
            >
          </div>
          <span class="text-sm">Sprint / Fast Move</span>
        </div>

        <div class="grid grid-cols-2 items-center gap-4">
          <div class="flex gap-1">
            <kbd class="px-2 py-1 bg-muted rounded border text-xs font-mono"
              >Space Bar</kbd
            >
            /
            <kbd class="px-2 py-1 bg-muted rounded border text-xs font-mono"
              >X</kbd
            >
          </div>
          <span class="text-sm">Up / Down</span>
        </div>
      </div>

      <DialogFooter>
        <Button @click="isShortcutsDialogOpen = false">Close</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

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

  <Dialog :open="isShareDialogOpen" @update:open="isShareDialogOpen = $event">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>Share</DialogTitle>
        <DialogDescription
          >Anyone who has this link will be able to view
          this.</DialogDescription
        >
      </DialogHeader>
      <div class="flex items-center space-x-2 pt-4">
        <div class="grid flex-1 gap-2">
          <Label for="link" class="sr-only">Link</Label>
          <Input id="link" :default-value="shareUrl" read-only class="h-9" />
        </div>
        <Button type="submit" size="sm" class="px-3" @click="copy(shareUrl)">
          <span class="sr-only">Copy</span>
          <Copy v-if="!copied" class="h-4 w-4" />
          <Check v-else class="h-4 w-4 text-green-500" />
        </Button>
      </div>
      <DialogFooter>
        <Button type="submit" @click="openDialog = !openDialog">Ok</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <TooltipProvider>
    <div
      class="h-16 border-b border-border px-6 flex items-center justify-between"
    >
      <!-- LEFT: Project info + save status -->
      <div id="left-menu" class="flex items-center gap-4">
        <Button
          class="cursor-pointer hover:bg-gray-100"
          variant="ghost"
          size="sm"
          @click="goToModels()"
        >
          <ArrowLeft class="w-4 h-4" />
        </Button>
        <Card class="px-3 py-1 bg-survey-surface">
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 bg-survey-accent rounded-full" />
            <span class="text-sm font-medium max-w-[50px] truncate">{{
              sceneStates?.modelInfo?.data.name ?? "Loading..."
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
              (sceneStates == undefined ||
                sceneStates.markedForCheck.value ||
                sceneStates.localVersion.value !==
                  sceneStates.lastSyncedVersion.value ||
                sceneStates.calibration.dirty ||
                sceneStates.markedFacesForCheck.value)
            "
          >
            <TooltipTrigger>
              <RefreshCcw
                class="animate-spin text-yellow-500"
                style="animation-direction: reverse"
            /></TooltipTrigger>
            <TooltipContent> Saving </TooltipContent>
          </Tooltip>
          <Tooltip v-else>
            <TooltipTrigger
              ><CloudCheck class="text-green-500"
            /></TooltipTrigger>
            <TooltipContent> Saved to Cloud </TooltipContent>
          </Tooltip>
        </div>
      </div>

      <!-- MIDDLE: Grouped controls -->
      <div id="middle-menu" class="flex items-center gap-2 shrink min-w-0">
        <!-- View Group: always visible -->
        <Button
          class="button-parent"
          size="sm"
          variant="outline"
          @click="toggleFullscreen"
        >
          <ClientOnly>
            <Tooltip>
              <TooltipTrigger><Maximize class="button-icon" /></TooltipTrigger>
              <TooltipContent>Fullscreen</TooltipContent>
            </Tooltip>
          </ClientOnly>
        </Button>

        <Button
          class="button-parent"
          size="sm"
          variant="outline"
          :class="isMapOpen ? 'bg-green-400! hover:bg-green-500!' : ''"
          @click="() => toggleMap()"
        >
          <Map class="button-icon" />
          <span class="ml-2 button-span-text">Map</span>
        </Button>

        <!-- Tools Group: workspace == 'me' only -->
        <template v-if="workspace == 'me'">
          <div class="h-6 w-px bg-border mx-2" />

          <Button
            class="button-parent"
            size="sm"
            variant="outline"
            :class="{ 'btn-camera-mode': currentPanel == 'camera' }"
            @click="() => toggleCameraPanel()"
          >
            <Camera class="button-icon" />
          </Button>

          <Button
            class="button-parent"
            size="sm"
            variant="outline"
            :class="{ 'btn-calibrating': currentPanel == 'calibration' }"
            @click="() => toggleCalibration()"
          >
            <RulerDimensionLine class="button-icon" />
            <span
              v-if="currentPanel != 'calibration'"
              class="ml-2 button-span-text"
              >Calibration</span
            >
            <span v-else class="ml-2 button-span-text">Calibrating...</span>
          </Button>

          <Button
            class="button-parent"
            size="sm"
            variant="outline"
            :class="{ 'btn-algo': currentPanel == 'algo' }"
            @click="() => toggleAlgoPanel()"
          >
            <Cpu class="tb-icon" />
            <span v-if="currentPanel != 'algo'" class="ml-2 button-span-text"
              >Algorithm</span
            >
            <span v-else class="ml-2 button-span-text">Algo Running...</span>
          </Button>
        </template>

        <!-- Optimization Group: candidate cameras pending review -->
        <template v-if="candidateEntries.length != 0">
          <div class="h-6 w-px bg-border mx-2" />

          <Button
            class="button-parent"
            size="sm"
            variant="outline"
            @click="confirmOptimizedCams"
          >
            <Check class="button-icon" />
            <span class="ml-2 button-span-text">Confirm</span>
          </Button>

          <Button
            class="button-parent"
            size="sm"
            variant="outline"
            @click="removeOptimizedCams"
          >
            <X class="button-icon" />
            <span class="ml-2 button-span-text">Cancel</span>
          </Button>
        </template>

        <!-- Workspace Actions Group -->
        <div class="h-6 w-px bg-border mx-2" />

        <Button
          v-if="workspace == 'me'"
          class="button-parent"
          size="sm"
          variant="outline"
          @click="saveModelToPublic()"
        >
          <Save class="button-icon" />
          <span class="ml-2 button-span-text">Publish</span>
        </Button>

        <Button
          v-if="workspace == 'me'"
          class="button-parent"
          size="sm"
          variant="outline"
          @click="
            goToModel();
            toggleCameraPanel();
          "
        >
          <LogOut class="button-icon" />
          <span class="ml-2 button-span-text">Exit Workspace</span>
        </Button>

        <template v-if="workspace == null">
          <Button
            v-if="sceneStates?.modelInfo.data.workspaceExists"
            class="button-parent"
            size="sm"
            variant="outline"
            @click="goToMyWorkspace()"
          >
            <PackageOpen class="button-icon" />
            <span class="ml-2 button-span-text">Open Workspace</span>
          </Button>
          <Button
            v-else
            class="button-parent"
            size="sm"
            variant="outline"
            @click="createWorkspace()"
          >
            <PackageOpen class="button-icon" />
            <span class="ml-2 button-span-text">Create Workspace</span>
          </Button>
        </template>
      </div>

      <!-- RIGHT: System controls -->
      <div
        id="right-menu"
        class="flex flex-row justify-center items-center gap-2"
      >
        <Tooltip v-if="workspace != null">
          <TooltipTrigger as-child>
            <Button
              class="button-parent"
              size="sm"
              variant="outline"
              @click="() => openShareDialog(true)"
            >
              <Share2 class="tb-icon" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Share workspace link</TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger as-child>
            <Button size="sm" variant="outline" @click="() => togglePanel()">
              <IndentIncrease v-if="isPanelOpen" class="button-icon" />
              <IndentDecrease v-else class="button-icon" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Toggle camera panel</TooltipContent>
        </Tooltip>

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
                @click="() => exportCamerasToJson(sceneStates?.cameras)"
              >
                <Upload class="button-icon" />
                Export
              </DropdownMenuItem>
              <DropdownMenuItem @click="toggleShortcutsDialog">
                <HelpCircle class="button-icon" />
                Shortcuts
              </DropdownMenuItem>
              <DropdownMenuItem @click="toggleSettingDialog">
                <Settings2 class="button-icon" />
                Setting
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator v-if="workspace == 'me'" />
            <DropdownMenuGroup>
              <DropdownMenuItem
                v-if="workspace == 'me'"
                @click="
                  deleteWorkspace();
                  toggleCameraPanel();
                "
              >
                <Trash2 class="button-icon text-red-500" />
                <span class="button-span-text text-red-500">
                  Delete Workspace
                </span>
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>

        <Setting3dDialog
          :open="isSettingDialogOpen"
          message="Update Sensitivity"
          @update:open="isSettingDialogOpen = $event"
        />

        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              size="sm"
              variant="outline"
              @click="lightDarkTheme.toggleTheme"
            >
              <Sun
                v-if="lightDarkTheme.theme.value == 'light'"
                class="button-icon"
              />
              <Moon v-else class="button-icon" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Toggle theme</TooltipContent>
        </Tooltip>

        <!-- User Actions -->
        <div class="flex items-center gap-2">
          <Tooltip>
            <TooltipTrigger as-child>
              <Button size="sm" variant="outline">
                <User class="button-icon" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{{ user?.username ?? "Account" }}</TooltipContent>
          </Tooltip>
        </div>
      </div>
    </div>
  </TooltipProvider>
</template>

<style lang="scss" scoped>
.button-parent {
  display: flex;
  align-items: center;
  min-width: 0;
  flex: 0 1 auto;
}
.button-icon {
  width: calc(4 * 0.25rem);
  height: calc(4 * 0.25rem);
  flex-shrink: 0;
}
.button-span-text {
  flex-shrink: 1;
  white-space: nowrap;
  display: block;
  text-overflow: ellipsis;
  overflow: hidden;
  min-width: 0;
}
Button {
  transition-property:
    color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}
.btn-camera-mode {
  background-color: #00c200 !important; /* blue-500 */
  color: white !important;
  border-color: #019701 !important;
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
.btn-algo {
  background-color: #3b82f6 !important; /* blue-500 */
  color: white !important;
  border-color: #2563eb !important;
}
.btn-algo:hover {
  background-color: #2563eb !important;
  opacity: 0.9;
}
</style>
