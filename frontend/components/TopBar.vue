<script setup lang="ts">
import Button from "./ui/button/Button.vue";
import Badge from "./ui/badge/Badge.vue";
import Card from "./ui/card/Card.vue";

import {
  RotateCcw,
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
} from "lucide-vue-next";

import { exportCamerasToJson } from "@/utils/exportScene";
import { importJsonToCameras } from "@/utils/importScene";
import { SCENE_STATES_KEY } from "@/components/3d/scene-states-provider/create-scene-states";

import Tooltip from "./ui/tooltip/Tooltip.vue";
import TooltipTrigger from "./ui/tooltip/TooltipTrigger.vue";
import TooltipContent from "./ui/tooltip/TooltipContent.vue";
import TooltipProvider from "./ui/tooltip/TooltipProvider.vue";
import {
  IS_PANEL_OPEN_KEY,
  MODEL_INFO_KEY,
  TOGGLE_PANEL_KEY,
} from "~/constants/state-keys";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const isPanelOpen = inject(IS_PANEL_OPEN_KEY);
const togglePanel = inject(TOGGLE_PANEL_KEY)!;

const route = useRoute();

const openDialog = ref(false);

const dialogTitle = ref("");
const dialogContent = ref("");

const lightDarkTheme = useLightDarkTheme();

async function saveModelToPublic() {
  const runtimeConfig = useRuntimeConfig();
  const resp = await fetch(
    `http://${runtimeConfig.public.externalBackendHost}/api/v1/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/merge`,
    { method: "POST", credentials: "include" },
  );

  if (!resp.ok) {
    console.error(resp);
    return;
  }

  const respJson: {
    noChanges?: boolean;
  } = await resp.json();

  openDialog.value = true;
  if (respJson.noChanges) {
    dialogTitle.value = "No changes";
    dialogContent.value = "There is no changes to be published";
  } else {
    dialogTitle.value = "Progress Saved";
    dialogContent.value = "";
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
    useState(MODEL_INFO_KEY, () => data);
    navigateTo(
      `/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me`,
    );
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

    useState(MODEL_INFO_KEY, () => undefined);
    navigateTo(
      `/projects/${route.params.projectId}/models/${route.params.modelId}`,
    );
  } catch (err) {
    console.error(err);
    showError({
      message: "Failed to create workspace",
    });
  }
}
</script>

<template>
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
            v-if="workspace != null && sceneStates.markedForCheck.size > 0"
          >
            <TooltipTrigger>
              <RefreshCcw class="animate-spin"
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
          :disabled="sceneStates.currentCamId.value == null"
          @click="sceneStates.cameraManagement.switchToSpectator()"
        >
          <RotateCcw class="button-icon" />
          <span class="ml-2 button-span-text"> Reset View </span>
        </Button>

        <Button size="sm" variant="outline">
          <Tooltip>
            <TooltipTrigger> <Maximize class="button-icon" /></TooltipTrigger>
            <TooltipContent> Fullscreen </TooltipContent>
          </Tooltip>
        </Button>

        <div class="h-6 w-px bg-border mx-2" />

        <Button
          v-if="workspace != null"
          size="sm"
          variant="outline"
          @click="saveModelToPublic()"
        >
          <Save class="button-icon" />
          <span class="ml-2 button-span-text"> Publish </span>
        </Button>

        <Button
          v-if="workspace != null"
          size="sm"
          variant="outline"
          @click="
            navigateTo(
              `/projects/${route.params.projectId}/models/${route.params.modelId}`,
            )
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
            @click="
              navigateTo(
                `/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me`,
              )
            "
          >
            <PackageOpen class="button-icon" />
            <span class="ml-2 button-span-text"> Open Workspace </span>
          </Button>
          <Button v-else size="sm" variant="outline" @click="createWorkspace()">
            <PackageOpen class="button-icon" />
            <span class="ml-2 button-span-text"> Create Workspace </span>
          </Button>
        </template>

        <Button size="sm" variant="outline" @click="() => togglePanel()">
          <IndentIncrease v-if="isPanelOpen" class="button-icon" />
          <IndentDecrease v-else class="button-icon" />
          <span class="ml-2 button-span-text"> Panel </span>
        </Button>
      </div>

      <div
        id="right-menu"
        class="flex flex-row justify-center items-center gap-2"
      >
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
  @apply transition-colors;
}
</style>
