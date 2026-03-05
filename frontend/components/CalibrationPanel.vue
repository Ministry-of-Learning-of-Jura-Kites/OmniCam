<script setup lang="ts">
import {
  IS_CALIBRATING_KEY,
  TOGGLE_CALIBRATION_KEY,
  SCENE_STATES_KEY,
} from "~/constants/state-keys";
import {
  Ruler,
  X,
  Check,
  Scaling,
  Pyramid,
  ArrowUpFromLine,
} from "lucide-vue-next";
const isCalibrating = inject(IS_CALIBRATING_KEY);
const toggleCalibration = inject(TOGGLE_CALIBRATION_KEY)!;
const realWorldSizeCm = ref(100);
const previousScaleFactor = ref(1);
const previousCalibrationGridScale = ref(1); // Default model height in cm
const sceneStates = inject(SCENE_STATES_KEY);
const calibrationGridScale = ref(
  realWorldSizeCm.value / 100 / sceneStates!.calibration.scale,
);
const canCancel = ref(false);

const confirmCalibration = () => {
  previousScaleFactor.value = sceneStates!.calibration.scale;
  previousCalibrationGridScale.value = calibrationGridScale!.value;
  canCancel.value = true;
  const realWorldSizeMeters = realWorldSizeCm.value / 100;
  const adjustment = realWorldSizeMeters / calibrationGridScale!.value;
  sceneStates!.calibration.scale = adjustment;
};
const cancelCalibration = () => {
  sceneStates!.calibration.scale = previousScaleFactor.value;
  calibrationGridScale!.value = previousCalibrationGridScale.value;
  canCancel.value = false;
};
const resetCalibration = () => {
  sceneStates!.calibration.scale = 1;
  calibrationGridScale!.value = 1;
  canCancel.value = false;
};
</script>

<template>
  <template v-if="isCalibrating">
    <div
      class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full shadow-lg"
    >
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-lg font-semibold flex items-center gap-2 text-primary">
          <Ruler class="h-5 w-5" />
          Calibration Manager
        </h2>
        <Button size="icon" variant="ghost" @click="toggleCalibration()">
          <X class="h-4 w-4" />
        </Button>
      </div>

      <div class="space-y-2">
        <Card class="gap-2">
          <CardHeader class="pb-1">
            <CardTitle class="text-sm font-medium flex items-center gap-2">
              <Pyramid class="h-4 w-4" />
              1. Set Real-World Distance
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-1">
            <div class="space-y-2">
              <Label>Grid Size in Real World (cm)</Label>
              <div class="flex items-center gap-2">
                <Input
                  v-model.number="realWorldSizeCm"
                  type="number"
                  placeholder="e.g., 100"
                  class="font-mono text-center"
                />
                <span class="text-sm font-medium">cm</span>
              </div>
              <p class="text-[10px] text-muted-foreground italic">
                * This red reference grid represents {{ realWorldSizeCm }} x
                {{ realWorldSizeCm }} cm in the physical world.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card class="gap-2">
          <CardHeader class="pb-1">
            <CardTitle class="text-sm font-medium flex items-center gap-2">
              <Scaling class="h-4 w-4" />
              2. Adjust Calibration Grid
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-3">
              <div class="flex justify-between text-xs">
                <span>Grid Visual Scale</span>
                <span class="font-mono"
                  >{{ calibrationGridScale!.toFixed(2) }}x</span
                >
              </div>
              <input
                v-model.number="calibrationGridScale"
                type="range"
                min="0.1"
                max="10"
                step="0.01"
                class="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-red-500"
              />
              <Input
                v-model.number="calibrationGridScale"
                type="number"
                step="0.1"
                class="h-8 text-xs"
              />
            </div>

            <p class="text-xs text-muted-foreground">
              Adjust the red grid's scale until it aligns with a known dimension
              on your 3D model (e.g., a tabletop or a wall edge).
            </p>
          </CardContent>
        </Card>

        <div class="pt-2 space-y-2">
          <Button class="w-full gap-2" @click="confirmCalibration">
            <Check class="h-4 w-4" />
            Calibration
          </Button>
          <div class="grid grid-cols-2 gap-2">
            <Button
              variant="outline"
              class="w-full text-destructive"
              :disabled="!canCancel"
              @click="cancelCalibration"
            >
              Cancel
            </Button>
            <Button
              variant="outline"
              class="w-full text-destructive"
              @click="resetCalibration"
            >
              Reset
            </Button>
          </div>

          <div class="relative flex py-1 items-center">
            <div class="grow border-t border-gray-400"></div>
            <span class="shrink mx-4 text-gray-400 text-sm"
              >Secondary Settings</span
            >
            <div class="grow border-t border-gray-400"></div>
          </div>
          <Card class="gap-2">
            <CardHeader class="pb-1">
              <CardTitle class="text-sm font-medium flex items-center gap-2">
                <ArrowUpFromLine class="h-4 w-4" />
                Adjust the Model's Height
              </CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="space-y-3">
                <div class="flex justify-between text-xs">
                  <span>Model's Height</span>
                  <span class="font-mono"
                    >{{
                      sceneStates!.calibration.heightOffset!.toFixed(2)
                    }}
                    m</span
                  >
                </div>
                <input
                  v-model.number="sceneStates!.calibration.heightOffset"
                  type="range"
                  min="-5"
                  max="5"
                  step="0.01"
                  class="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-red-500"
                />
                <Input
                  v-model.number="sceneStates!.calibration.heightOffset"
                  type="number"
                  step="0.1"
                  class="h-8 text-xs"
                />
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  </template>
</template>
