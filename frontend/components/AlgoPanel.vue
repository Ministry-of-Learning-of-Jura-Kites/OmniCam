<script setup lang="ts">
import { computed, inject, ref } from "vue";
import { SCENE_STATES_KEY } from "@/constants/state-keys";

const sceneStates = inject(SCENE_STATES_KEY)!;

const numCameras = ref(3);
const isOptimizing = ref(false);

const selectedCount = computed(
  () => sceneStates.selectedCoverageFaces.value.length,
);

const startAreaSelection = () => {
  sceneStates.setSelectionMode("coverage-area");
};

const stopAreaSelection = () => {
  sceneStates.setSelectionMode("none");
};

const clearAreas = () => {
  sceneStates.clearCoverageFaces();
};
</script>

<template>
  <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        Predictive Camera Placement
      </h2>
    </div>

    <div class="rounded-xl border border-border p-3 mb-4 bg-muted/20">
      <p class="text-sm font-medium mb-1">Target Area Selection Mode</p>

      <div class="mt-3 grid grid-cols-2 gap-2">
        <button
          class="py-2 bg-blue-600 hover:bg-blue-700 rounded-lg font-medium transition text-white"
          @click="startAreaSelection"
        >
          Start Selecting
        </button>
        <button
          class="py-2 bg-zinc-700 hover:bg-zinc-600 rounded-lg font-medium transition text-white"
          @click="stopAreaSelection"
        >
          Stop
        </button>
      </div>

      <button
        class="w-full mt-2 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg font-medium transition"
        @click="clearAreas"
      >
        Clear Selected Areas
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
        v-for="(face, i) in sceneStates.selectedCoverageFaces.value"
        :key="face.id"
      >
        <div class="font-medium mb-1">Area {{ i + 1 }}</div>

        <div class="opacity-60 mt-1">
          center: ({{ face.center[0].toFixed(2) }},
          {{ face.center[1].toFixed(2) }}, {{ face.center[2].toFixed(2) }})
        </div>
      </div>
    </div>

    <button
      :disabled="selectedCount === 0 || isOptimizing"
      class="w-full py-3.5 bg-emerald-600 hover:bg-emerald-700 disabled:bg-zinc-700 rounded-xl font-semibold text-lg transition"
    >
      {{ isOptimizing ? "Optimizing..." : "Run Optimization" }}
    </button>
  </div>
</template>
