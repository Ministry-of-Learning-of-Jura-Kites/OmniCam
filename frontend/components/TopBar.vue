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
} from "lucide-vue-next";

import { exportCamerasToJson } from "@/utils/exportScene";
import { importJsonToCameras } from "@/utils/importScene";
import { SCENE_STATES_KEY } from "@/components/3d/scene-states-provider/create-scene-states";

import Tooltip from "./ui/tooltip/Tooltip.vue";
import TooltipTrigger from "./ui/tooltip/TooltipTrigger.vue";
import TooltipContent from "./ui/tooltip/TooltipContent.vue";
import TooltipProvider from "./ui/tooltip/TooltipProvider.vue";
import { MODEL_INFO_KEY } from "~/constants/state-keys";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const isPanelOpen = inject("isPanelOpen") as Ref<boolean>;
const togglePanel = inject("togglePanel") as () => void;

const route = useRoute();

const openDialog = ref(false);

const dialogTitle = ref("");
const dialogContent = ref("");

async function saveModelToPublic() {
  const runtimeConfig = useRuntimeConfig();
  const resp = await fetch(
    `http://${runtimeConfig.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/merge`,
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
      importJsonToCameras(sceneStates.cameras, text);
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
      `http://${runtimeConfig.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${route.params.projectId}/models/${route.params.modelId}/workspaces/me`,
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
      class="h-16 bg-card border-b border-border px-6 flex items-center justify-between"
    >
      <!-- Project Info -->
      <div class="flex items-center gap-4">
        <Card class="px-3 py-1 bg-survey-surface">
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 bg-survey-accent rounded-full" />
            <span class="text-sm font-medium">Survey Project 01</span>
            <Badge variant="secondary" class="ml-2">
              {{ props.workspace == null ? "Public" : "Workspace" }}
            </Badge>
          </div>
        </Card>
        <div class="flex items-center justify-center h-4 w-4">
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
      <div class="flex items-center gap-2">
        <Button
          size="sm"
          variant="outline"
          :disabled="sceneStates.currentCamId.value == null"
          @click="sceneStates.cameraManagement.switchToSpectator()"
        >
          <RotateCcw class="h-4 w-4 mr-2" />
          Reset View
        </Button>

        <Button size="sm" variant="outline">
          <Maximize class="h-4 w-4 mr-2" />
          Fullscreen
        </Button>

        <div class="h-6 w-px bg-border mx-2" />

        <Button
          v-if="workspace != null"
          size="sm"
          variant="outline"
          @click="saveModelToPublic()"
        >
          <Save class="h-4 w-4 mr-2" />
          Publish
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
          <LogOut class="h-4 w-4 mr-2" />
          Exit Workspace
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
            <PackageOpen class="h-4 w-4 mr-2" />
            Open Workspace
          </Button>
          <Button v-else size="sm" variant="outline" @click="createWorkspace()">
            <PackageOpen class="h-4 w-4 mr-2" />
            Create Workspace
          </Button>
        </template>

        <Button size="sm" variant="outline" @click="() => openFileDialog()">
          <Upload class="h-4 w-4 mr-2" />
          Import
        </Button>

        <Button
          size="sm"
          variant="outline"
          @click="() => exportCamerasToJson(sceneStates.cameras)"
        >
          <Download class="h-4 w-4 mr-2" />
          Export
        </Button>

        <Button size="sm" variant="outline" @click="() => togglePanel()">
          <IndentIncrease v-if="isPanelOpen" class="h-4 w-4 mr-2" />
          <IndentDecrease v-else class="h-4 w-4 mr-2" />
          Panel
        </Button>
      </div>

      <!-- User Actions -->
      <div class="flex items-center gap-2">
        <Button size="sm" variant="ghost">
          <User class="h-4 w-4 mr-2" />
          Profile
        </Button>
        <Button size="sm" variant="ghost">
          <LogOut class="h-4 w-4" />
        </Button>
      </div>
    </div>
  </TooltipProvider>
</template>
