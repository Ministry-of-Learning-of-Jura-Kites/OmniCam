<script setup lang="ts">
import { computed, inject, ref } from "vue";
import { SCENE_STATES_KEY } from "@/constants/state-keys";
import { Eye, EyeOff, Plus, Trash } from "lucide-vue-next";
import type { Camerapreset } from "./CameraPanel.vue";
import LazyCameraSpawnDialog from "./dialog/CameraSpawnDialog.vue";
import Card from "./ui/card/Card.vue";
import type { CameraConfig } from "~/messages/protobufs/optimization";
import { v4 as uuidv4 } from "uuid";

const isCameraSpawnDialogOpen = ref(false);

const sceneStates = inject(SCENE_STATES_KEY)!;

const cameraConfigs = reactive<CameraConfig[]>([]);

const errorMsg = ref<string | null>(null);

const faceEntries = computed(() =>
  Object.entries(sceneStates.facesManagement.faces),
);

const isAllSelected = computed(() => {
  return (
    faceEntries.value.length > 0 &&
    selectedIds.size === faceEntries.value.length
  );
});

const selectedIds = reactive(new Set<string>());

const facesCount = computed(
  () => Object.keys(sceneStates.facesManagement.faces).length,
);

const selectedSize = computed(() => {
  if (isAllSelected.value) {
    return facesCount.value;
  }
  return selectedIds.size;
});

const toggleAreaSelection = () => {
  if (sceneStates.selectionMode.value == "none") {
    sceneStates.facesManagement.setMode("coverage-area");
  } else {
    sceneStates.facesManagement.setMode("none");
  }
};

const clearAreas = () => {
  sceneStates.facesManagement.clear();
};

const deleteArea = (faceId: string) => {
  sceneStates.facesManagement.remove(faceId);
  selectedIds.delete(faceId);
};

const updateAreaColor = (faceId: string, event: Event) => {
  const color = (event.target as HTMLInputElement | null)?.value;
  if (!color) return;
  sceneStates.facesManagement.updateColor(faceId, color);
};

const getAreaColor = (color?: string) => color ?? "#22ff88";
const isAllCoverageHidden = computed(
  () => sceneStates.facesManagement.isAllHidden.value,
);

const toggleAllAreasVisibility = () => {
  sceneStates.facesManagement.toggleAllHidden();
};

const toggleAreaVisibility = (faceId: string) => {
  sceneStates.facesManagement.toggleFaceHidden(faceId);
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
  sceneStates.optimization?.requestOptimize(
    faceEntries.value.filter(([id, _face]) => selectedIds.has(id)),
    sceneStates.calibration.scale,
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
        <h2 class="text-lg font-semibold flex items-center gap-2">
          Predictive Camera Placement
        </h2>
      </div>

      <div class="rounded-xl border border-border p-3 mb-4 bg-muted/20">
        <p class="text-sm font-medium mb-1">Target Area Selection Mode</p>

        <div class="mt-3 grid grid-cols-1 gap-2">
          <button
            class="py-2 bg-blue-600 hover:bg-blue-700 rounded-lg font-medium transition text-white"
            :class="{
              'bg-red-600': sceneStates.selectionMode.value !== 'none',
              'hover:bg-red-700': sceneStates.selectionMode.value !== 'none',
            }"
            @click="toggleAreaSelection"
          >
            {{
              sceneStates.selectionMode.value == "none"
                ? "Start Selecting"
                : "Stop Selecting"
            }}
          </button>
        </div>

        <button
          class="w-full mt-2 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg font-medium transition"
          @click="clearAreas"
        >
          Clear Selected Areas
        </button>
        <button
          class="w-full mt-2 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg font-medium transition"
          @click="toggleAllAreasVisibility"
        >
          {{ isAllCoverageHidden ? "Show All Areas" : "Hide All Areas" }}
        </button>
      </div>

      <div class="rounded-xl border border-border p-3 mb-4 bg-muted/20">
        <div class="flex flex-row justify-between items-center">
          <span class="block text-sm font-medium">Cameras To Use</span>
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
            sceneStates.facesManagement.faces,
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
        class="w-full text-white bg-emerald-400! hover:bg-emerald-500! disabled:bg-red-800 text-lg transition"
        @click="submit()"
      >
        {{
          {
            idle: "Run Optimization",
            sending: "Sending..",
            optimizing: "Optimizing..",
          }[sceneStates.optimization!.submitStatus.value]
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
  gap: calc(var(--spacing) * 6);
}
</style>
