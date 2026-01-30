<script setup lang="ts">
import {
  IS_CALIBRATING_KEY,
  TOGGLE_CALIBRATION_KEY,
} from "~/constants/state-keys";
import { RulerDimensionLine, X } from "lucide-vue-next";

const isCalibrating = inject(IS_CALIBRATING_KEY);
const toggleCalibration = inject(TOGGLE_CALIBRATION_KEY)!;
</script>

<template v-if="isCalibrating">
  <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        <RulerDimensionLine class="button-icon" />
        Calibration Manager
      </h2>
      <Button size="sm" @click="toggleCalibration()">
        <X class="h-4 w-4" />
      </Button>
    </div>
  </div>
</template>
