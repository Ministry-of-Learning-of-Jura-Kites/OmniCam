<script setup lang="ts">
import { computed, inject, ref } from "vue";
import { SCENE_STATES_KEY } from "@/constants/state-keys";
import { Eye, EyeOff, Trash } from "lucide-vue-next";

const sceneStates = inject(SCENE_STATES_KEY)!;

const numCameras = ref(3);
const isOptimizing = ref(false);

const selectedCount = computed(
  () => Object.keys(sceneStates.facesManagement.faces).length,
);

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
</script>

<template>
  <TooltipProvider>
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
        <label class="block text-sm font-medium mb-2">Number of Cameras</label>
        <input
          v-model.number="numCameras"
          type="number"
          min="1"
          max="10"
          step="1"
          class="w-full rounded-md bg-background border border-border px-3 py-2 text-sm"
        />
      </div>

      <div v-if="selectedCount > 0" class="mb-6">
        <p class="text-sm text-gray-400 mb-2">
          Selected Target Areas ({{ selectedCount }})
        </p>

        <div
          v-for="[i, face] of Object.entries(sceneStates.facesManagement.faces)"
          :key="face.id"
          class="mb-3 rounded-lg border border-border bg-muted/20 p-3"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="font-medium">
              Area {{ i + 1 }}
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
                    @click="toggleAreaVisibility(face.id)"
                  >
                    <Eye v-if="face.hidden" class="h-3 w-3" />
                    <EyeOff v-else class="h-3 w-3" />
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
                    @click="deleteArea(face.id)"
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
              @input="updateAreaColor(face.id, $event)"
            />
            <span class="text-xs opacity-60">
              {{ getAreaColor(face.color) }}
            </span>
          </div>
        </div>
      </div>

      <Button
        :disabled="selectedCount === 0 || isOptimizing"
        variant="default"
        class="w-full text-black bg-emerald-400! hover:bg-emerald-500! disabled:bg-red-800 text-lg transition"
      >
        {{ isOptimizing ? "Optimizing..." : "Run Optimization" }}
      </Button>
    </div>
  </TooltipProvider>
</template>
