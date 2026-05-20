<script setup lang="ts">
import { computed, inject } from "vue";
import { PANEL_KEY, SCENE_STATES_KEY } from "~/constants/state-keys";
import {
  Ruler,
  Trash2,
  MousePointerClick,
  Scale3D,
  CircleDot,
  LineChart,
} from "lucide-vue-next";

// UI Components
import Button from "./ui/button/Button.vue";
import Card from "./ui/card/Card.vue";
import CardHeader from "./ui/card/CardHeader.vue";
import CardTitle from "./ui/card/CardTitle.vue";
import CardContent from "./ui/card/CardContent.vue";

const sceneStates = inject(SCENE_STATES_KEY)!;
const { currentToolMode } = inject(PANEL_KEY)!;

const measurementLines = sceneStates.value!.measurement.lines;
const pendingPoint = computed(
  () => sceneStates.value!.measurement.draftStartPoint,
);
watch(
  () => measurementLines.length,
  (newLength) => {
    console.log("Measurement lines count:", newLength);
    console.log("Current lines:", measurementLines);
  },
);
const totalMeasurements = computed(() => measurementLines.length);

function clearMeasurements() {
  if (sceneStates.value?.measurement) {
    sceneStates.value.measurement.lines.splice(0);
    sceneStates.value.measurement.draftStartPoint = null;
  }
}

function formatDistance(distance: number) {
  return `${distance.toFixed(2)} m`;
}
</script>

<template>
  <template v-if="currentToolMode === 'measurement'">
    <div
      class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full shadow-lg"
    >
      <!-- HEADER -->
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-lg font-semibold flex items-center gap-2 text-primary">
          <Ruler class="h-5 w-5" />
          Measurement Tool
        </h2>
      </div>

      <div class="space-y-2">
        <!-- STATUS -->
        <Card class="gap-2">
          <div class="px-4 py-2 flex justify-between items-center">
            <CardTitle class="text-sm font-medium flex items-center gap-2">
              <MousePointerClick class="h-4 w-4" />
              Measurement Status
            </CardTitle>
          </div>

          <CardContent class="px-4 space-y-3">
            <div class="flex items-center justify-between text-sm">
              <span class="text-muted-foreground">Waiting For</span>
              <span class="font-medium">
                {{ pendingPoint ? "Second Point" : "First Point" }}
              </span>
            </div>
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
                  >alt</kbd
                >
              </span>
              <span>+ Click to define point</span>
            </p>
          </CardContent>
        </Card>

        <!-- CURRENT SCALE -->
        <Card class="gap-2">
          <div class="px-4 py-2 flex justify-between items-center">
            <CardTitle class="text-sm font-medium flex items-center gap-2">
              <Scale3D class="h-4 w-4" />
              Calibration Scale
            </CardTitle>
          </div>

          <CardContent class="px-4">
            <div class="flex items-center gap-4">
              <div class="flex flex-col">
                <span
                  class="text-[10px] uppercase text-muted-foreground font-semibold"
                >
                  Virtual
                </span>
                <span class="text-2xl font-mono font-bold">1m</span>
              </div>

              <div class="h-8 w-px bg-border rotate-20" />

              <div class="flex flex-col">
                <span
                  class="text-[10px] uppercase text-muted-foreground font-semibold"
                >
                  Real World
                </span>
                <span class="text-2xl font-mono font-bold text-primary">
                  {{ sceneStates!.calibration?.scale ?? 1 }}m
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- SUMMARY -->
        <Card class="gap-2">
          <CardHeader class="pb-1">
            <CardTitle class="text-sm font-medium flex items-center gap-2">
              <LineChart class="h-4 w-4" />
              Measurement Summary
            </CardTitle>
          </CardHeader>

          <CardContent class="space-y-3">
            <div class="flex justify-between text-sm">
              <span>Total Measurements</span>
              <span class="font-mono">
                {{ totalMeasurements }}
              </span>
            </div>

            <div class="flex justify-between text-sm">
              <span>Pending Point</span>
              <span class="font-mono">
                {{ pendingPoint ? "Yes" : "No" }}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card class="gap-2">
          <CardHeader class="pb-1">
            <div class="flex items-center justify-between">
              <CardTitle class="text-sm font-medium flex items-center gap-2">
                <CircleDot class="h-4 w-4" />
                Measurements
              </CardTitle>
              <Button
                variant="destructive"
                size="sm"
                class="h-8 px-3"
                :disabled="measurementLines.length === 0"
                @click="clearMeasurements"
              >
                <Trash2 class="h-4 w-4 mr-1" />
                Clear All
              </Button>
            </div>
          </CardHeader>

          <CardContent>
            <div
              v-if="measurementLines.length === 0"
              class="text-xs text-muted-foreground italic"
            >
              No measurements yet.
            </div>

            <div v-else class="space-y-2 max-h-64 overflow-y-auto">
              <div
                v-for="(measurement, index) in measurementLines"
                :key="measurement.id"
                class="border rounded-md p-3 bg-muted/40"
              >
                <div class="flex justify-between items-start">
                  <div class="flex flex-col">
                    <span class="text-xs text-muted-foreground">
                      Measurement {{ index + 1 }}
                    </span>

                    <span class="font-mono font-semibold">
                      {{ formatDistance(measurement.realDistance) }}
                    </span>

                    <span class="text-xs text-muted-foreground mt-1">
                      Virtual:
                      {{ formatDistance(measurement.virtualDistance) }}
                    </span>
                  </div>

                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 text-destructive hover:text-destructive"
                    @click="sceneStates!.measurement.lines.splice(index, 1)"
                  >
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </template>
</template>
