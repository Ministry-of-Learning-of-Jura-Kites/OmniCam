<script setup lang="ts">
import { computed, inject, ref } from "vue";
import { SCENE_STATES_KEY } from "@/constants/state-keys";
import {
  Eye,
  EyeOff,
  Plus,
  Trash,
  SlidersHorizontal,
  Crosshair,
  Video,
} from "lucide-vue-next";
import type { Camerapreset } from "./CameraPanel.vue";
import LazyCameraSpawnDialog from "./dialog/CameraSpawnDialog.vue";
import Card from "./ui/card/Card.vue";
import type { CameraConfig } from "~/messages/protobufs/optimization";
import { v4 as uuidv4 } from "uuid";

const isCameraSpawnDialogOpen = ref(false);

const sceneStates = inject(SCENE_STATES_KEY);

const cameraConfigs = reactive<CameraConfig[]>([]);

const errorMsg = ref<string | null>(null);

const faceEntries = computed(() =>
  Object.entries(sceneStates?.value?.facesManagement.faces ?? {}),
);

const isAllSelected = computed(() => {
  return (
    faceEntries.value.length > 0 &&
    selectedIds.size === faceEntries.value.length
  );
});

const selectedIds = reactive(new Set<string>());

const facesCount = computed(
  () => Object.keys(sceneStates?.value?.facesManagement.faces ?? {}).length,
);

const selectedSize = computed(() => {
  if (isAllSelected.value) {
    return facesCount.value;
  }
  return selectedIds.size;
});

const toggleAreaSelection = () => {
  if (sceneStates == undefined) {
    return;
  }
  if (sceneStates.value!.selectionMode.value == "none") {
    sceneStates.value!.facesManagement.setMode("coverage-area");
  } else {
    sceneStates.value!.facesManagement.setMode("none");
  }
};

const clearAreas = () => {
  if (sceneStates == undefined) {
    return;
  }
  sceneStates.value!.facesManagement.clear();
};

const deleteArea = (faceId: string) => {
  if (sceneStates == undefined) {
    return;
  }
  sceneStates.value!.facesManagement.remove(faceId);
  selectedIds.delete(faceId);
};

const updateAreaColor = (faceId: string, event: Event) => {
  if (sceneStates == undefined) {
    return;
  }
  const color = (event.target as HTMLInputElement | null)?.value;
  if (!color) return;
  sceneStates.value!.facesManagement.updateColor(faceId, color);
};

const getAreaColor = (color?: string) => color ?? "#22ff88";
const isAllCoverageHidden = computed(
  () => sceneStates?.value?.facesManagement.isAllHidden.value ?? true,
);

const toggleAllAreasVisibility = () => {
  if (sceneStates == undefined) {
    return;
  }
  sceneStates.value!.facesManagement.toggleAllHidden();
};

const toggleAreaVisibility = (faceId: string) => {
  if (sceneStates == undefined) {
    return;
  }
  sceneStates.value!.facesManagement.toggleFaceHidden(faceId);
};

function handleAddCameraConfig(preset: Camerapreset) {
  cameraConfigs.push({
    id: uuidv4(),
    name: `${preset.vendor} ${preset.camera} ${preset.sensor_name}`,
    fov: Number(preset.fov),
    widthRes: Number(preset.res_w),
    heightRes: Number(preset.res_h),
    amount: 1,
  });
}

function removeCamConfig(camConfig: CameraConfig) {
  const index = cameraConfigs.indexOf(camConfig);
  if (index > -1) {
    cameraConfigs.splice(index, 1);
  }
}

function toggleIdSelection(id: string) {
  if (selectedIds.has(id)) {
    selectedIds.delete(id);
  } else {
    selectedIds.add(id);
  }
}
function bulkDelete() {
  selectedIds.forEach((id) => deleteArea(id));
  selectedIds.clear(); // Clear selection after deletion
}

function bulkToggleVisibility() {
  selectedIds.forEach((id) => toggleAreaVisibility(id));
}

function toggleSelectAll(checked: boolean | "indetermediate") {
  if (checked) {
    faceEntries.value.forEach(([id]) => selectedIds.add(id));
  } else {
    selectedIds.clear();
  }
}

function submit() {
  if (sceneStates == undefined) {
    return;
  }
  sceneStates.value!.optimization?.requestOptimize(
    faceEntries.value.filter(([id, _face]) => selectedIds.has(id)),
    sceneStates.value!.calibration.scale,
    cameraConfigs,
  );
}
</script>

<template>
  <TooltipProvider>
    <LazyCameraSpawnDialog
      v-model="isCameraSpawnDialogOpen"
      :on-confirm="handleAddCameraConfig"
    />
    <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold flex items-center gap-2 text-primary">
          <SlidersHorizontal class="h-5 w-5" />
          Predictive Camera Placement
        </h2>
      </div>

      <div class="rounded-xl border border-border p-3 mb-4 bg-muted/20">
        <p class="text-sm font-medium mb-1 flex items-center gap-2">
          <Crosshair class="h-4 w-4" />
          Target Area Selection Mode
        </p>

        <div class="mt-3 grid grid-cols-1 gap-2">
          <Button
            size="sm"
            variant="outline"
            class="w-full"
            :class="
              sceneStates?.selectionMode.value !== 'none'
                ? 'bg-red-500! hover:bg-red-700! text-white border-red-600!'
                : 'bg-blue-500! hover:bg-blue-700! text-white border-blue-600!'
            "
            @click="toggleAreaSelection"
          >
            {{
              sceneStates?.selectionMode.value !== "coverage-area"
                ? "Start Selecting"
                : "Stop Selecting"
            }}
          </Button>

          <p
            class="text-[11px] text-muted-foreground text-center mt-1 flex items-center justify-center gap-1"
          >
            <span class="flex items-center gap-0.5">
              <kbd
                class="px-1 py-0.5 rounded border bg-background font-sans text-xs"
                >Ctrl</kbd
              >
              <span class="text-muted-foreground">/</span>
              <kbd
                class="px-1 py-0.5 rounded border bg-background font-sans text-xs"
                >⌘</kbd
              >
            </span>
            <span>+ Click to define corners</span>
          </p>
        </div>

        <div class="grid grid-cols-2 gap-2 mt-2">
          <Button
            size="sm"
            variant="outline"
            class="w-full"
            @click="clearAreas"
          >
            <Trash class="h-4 w-4 mr-1" /> Clear Areas
          </Button>
          <Button
            size="sm"
            variant="outline"
            class="w-full"
            @click="toggleAllAreasVisibility"
          >
            <EyeOff v-if="!isAllCoverageHidden" class="h-4 w-4 mr-1" />
            <Eye v-else class="h-4 w-4 mr-1" />
            {{ isAllCoverageHidden ? "Show All" : "Hide All" }}
          </Button>
        </div>
      </div>

      <div class="rounded-xl border border-border p-3 mb-4 bg-muted/20">
        <div class="flex flex-row justify-between items-center">
          <span class="block text-sm font-medium flex items-center gap-2">
            <Video class="h-4 w-4" /> Cameras To Use
          </span>
          <Button
            size="sm"
            variant="outline"
            @click="isCameraSpawnDialogOpen = true"
          >
            <Plus />
          </Button>
        </div>
        <div class="flex flex-col px-2">
          <Card
            v-for="camConfig of cameraConfigs"
            :key="camConfig.id"
            class="px-2 mt-3 gap-3"
          >
            <div class="cam-config-field">
              <span>Name:</span><Input v-model.number="camConfig.name" />
            </div>
            <div class="cam-config-field">
              <span>Vertical FOV:</span><Input v-model.number="camConfig.fov" />
            </div>
            <div class="cam-config-field">
              <span>Width Resolution:</span>
              <Input v-model.number="camConfig.widthRes" min="1" />
            </div>
            <div class="cam-config-field">
              <span>Height Resolution:</span>
              <Input v-model.number="camConfig.heightRes" min="1" />
            </div>
            <div class="cam-config-field">
              <span>Amount:</span>
              <Input v-model.number="camConfig.amount" min="1" />
            </div>
            <Button variant="destructive" @click="removeCamConfig(camConfig)">
              <Trash />
            </Button>
          </Card>
        </div>
      </div>

      <Card v-if="facesCount > 0" class="mb-6 px-2 gap-2">
        <div class="flex items-center justify-between mb-4 px-2">
          <div class="flex items-center gap-2">
            <Checkbox
              :model-value="isAllSelected"
              @update:model-value="(v) => toggleSelectAll(v === true)"
            />
            <p class="text-sm text-gray-400">
              Selected Target Areas ({{ selectedSize }} / {{ facesCount }})
            </p>
          </div>

          <div
            v-if="selectedIds.size > 0"
            class="flex gap-2 animate-in fade-in zoom-in duration-200"
          >
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  size="sm"
                  variant="outline"
                  class="px-2"
                  @click="bulkToggleVisibility"
                >
                  <Eye class="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                Toggle Visibility ({{ selectedIds.size }})
              </TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  size="sm"
                  variant="destructive"
                  class="px-2"
                  @click="bulkDelete"
                >
                  <Trash class="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                Delete Selected ({{ selectedIds.size }})
              </TooltipContent>
            </Tooltip>
          </div>
        </div>

        <div
          v-for="[id, face] of Object.entries(
            sceneStates?.facesManagement.faces ?? {},
          )"
          :key="id"
          class="mb-3 rounded-lg border border-border bg-muted/20 p-3"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="font-medium min-w-0 wrap-break-words">
              <Checkbox
                :model-value="selectedIds.has(id)"
                @update:model-value="toggleIdSelection(id)"
              />
              Name: {{ face.name }}
              <span
                v-if="face.hidden"
                class="ml-2 text-xs px-2 py-0.5 rounded bg-zinc-700 text-zinc-200"
              >
                Hidden
              </span>
            </div>

            <div class="flex gap-2">
              <Tooltip>
                <TooltipTrigger>
                  <Button
                    size="sm"
                    variant="outline"
                    class="px-2 py-1 text-sm rounded-md hover:bg-zinc-600 text-white transition"
                    @click="toggleAreaVisibility(id)"
                  >
                    <EyeOff v-if="face.hidden" class="h-3 w-3" />
                    <Eye v-else class="h-3 w-3" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  {{ face.hidden ? "Show" : "Hide" }}
                </TooltipContent>
              </Tooltip>

              <Tooltip>
                <TooltipTrigger>
                  <Button
                    size="sm"
                    variant="destructive"
                    class="px-2 py-1 text-sm rounded-md hover:bg-zinc-600 text-white transition"
                    @click="deleteArea(id)"
                  >
                    <Trash class="h-3 w-3" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent> Delete </TooltipContent>
              </Tooltip>
            </div>
          </div>

          <div class="opacity-60 mt-2 text-sm">
            center: ({{ face.center?.[0].toFixed(2) }},
            {{ face.center?.[1].toFixed(2) }},
            {{ face.center?.[2].toFixed(2) }})
          </div>

          <div class="mt-3 flex items-center gap-3">
            <label class="text-sm opacity-80">Area Color</label>
            <input
              type="color"
              :value="getAreaColor(face.color)"
              class="h-10 w-16 cursor-pointer rounded border border-border bg-transparent"
              @input="updateAreaColor(id, $event)"
            />
            <span class="text-xs opacity-60">
              {{ getAreaColor(face.color) }}
            </span>
          </div>
        </div>
      </Card>

      <Button
        :disabled="
          facesCount === 0 || cameraConfigs.length == 0 || selectedIds.size == 0
        "
        variant="default"
        class="w-full text-white bg-emerald-400! hover:bg-emerald-500! disabled:bg-red-800 transition"
        @click="submit()"
      >
        {{
          {
            idle: "Run Optimization",
            sending: "Sending..",
            optimizing: "Optimizing..",
          }[sceneStates?.optimization?.submitStatus?.value ?? "idle"]
        }}
      </Button>
    </div>
  </TooltipProvider>
  <span v-if="errorMsg">{{ errorMsg }}</span>
</template>

<style lang="scss" scoped>
.cam-config-field {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.875rem;
}
</style>
