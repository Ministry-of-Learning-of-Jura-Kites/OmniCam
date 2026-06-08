<script setup lang="ts">
import { computed, inject, ref, watch, onMounted } from "vue";
import { PANEL_KEY, SCENE_STATES_KEY } from "~/constants/state-keys";
import { useWorkspaceApi } from "~/composables/api/use-workspace-api.js";
import { v4 as uuidv4 } from "uuid";
import { Vector3 } from "three";

import type { PopulationGroup } from "~/types/simulation";

import {
  X,
  Plus,
  CircleDot,
  Users,
  UsersRound,
  UserRound,
  LogIn,
  LogOut,
  Ruler,
  Gauge,
  Hash,
  Trash2,
  Save,
  Play,
  Map,
  Minus,
  Spline,
  Route as RouteIcon,
} from "lucide-vue-next";

const sceneStates = inject(SCENE_STATES_KEY)!;
const { currentPanel } = inject(PANEL_KEY)!;

const route = useRoute();
const runtimeConfig = useRuntimeConfig();

const projectId = route.params.projectId as string;
const modelId = route.params.modelId as string;
const workspaceId = route.params.workspaceId as string;

const activeRouteId = ref<string>("");

const savedSimulation = ref("");
const isDirty = ref<boolean>(false);

const workspaceApi = useWorkspaceApi(
  projectId,
  modelId,
  workspaceId,
  runtimeConfig,
);

const isSimulationRunning = computed(
  () => sceneStates.value?.simulationState.value === "running",
);

function runSimulation() {
  const state = sceneStates.value!.simulationState.value;

  if (state === "running") {
    sceneStates.value!.simulationState.value = "idle";
    return;
  }

  sceneStates.value!.simulationState.value = "running";
}

const simulation = computed(() => {
  if (!sceneStates.value) {
    return { areas: [], populationGroups: [], routes: [] };
  }
  const areas = Object.entries(sceneStates.value.facesManagement.faces)
    .filter(([, face]) => face.type === "simulation")
    .map(([id, face]) => ({
      id,
      name: face.name,
      color: face.color,
      kind: face.kind,
      points: face.points,
    }));

  return {
    areas,
    populationGroups: sceneStates.value.simulation?.populationGroups ?? [],
    routes: sceneStates.value.simulation?.routes ?? [],
  };
});

const simulationAreaCount = computed(() => simulation.value.areas.length);

const startAreas = computed(() =>
  simulation.value.areas
    .filter((a) => a.kind === "start")
    .slice()
    .sort((a, b) => extractNumber(a.name) - extractNumber(b.name)),
);

const endAreas = computed(() =>
  simulation.value.areas
    .filter((a) => a.kind === "end")
    .slice()
    .sort((a, b) => extractNumber(a.name) - extractNumber(b.name)),
);

// ====================== SELECTION MODE ======================
type SimulationSelectMode = "none" | "start" | "end";
const simulationSelectMode = ref<SimulationSelectMode>("none");

const isLoadingSimulation = ref(false);

function extractNumber(name: string | undefined) {
  const match = name?.match(/(\d+)/);
  return match ? Number(match[1]) : 0;
}

function syncRoutes() {
  if (!sceneStates.value) return;
  if (isLoadingSimulation.value) return; // ← guard against load-time triggering
  if (startAreas.value.length === 0 || endAreas.value.length === 0) return;

  const routes = sceneStates.value.simulation.routes;
  const existingRouteIds = new Set(routes.map((r) => r.id));

  // Add missing routes for every start×end combination
  for (const start of startAreas.value.sort(
    (a, b) => extractNumber(a.name) - extractNumber(b.name),
  )) {
    for (const end of endAreas.value.sort(
      (a, b) => extractNumber(a.name) - extractNumber(b.name),
    )) {
      const routeId = `route-${start.id}-${end.id}`;
      if (!existingRouteIds.has(routeId)) {
        routes.push({
          id: routeId,
          name: `${start.name || "Start"} → ${end.name || "End"}`,
          startAreaId: start.id,
          endAreaId: end.id,
          segments: [],
        });
      }
    }
  }

  routes.sort((a, b) => {
    const startA = startAreas.value.find((x) => x.id === a.startAreaId);
    const startB = startAreas.value.find((x) => x.id === b.startAreaId);

    const endA = endAreas.value.find((x) => x.id === a.endAreaId);
    const endB = endAreas.value.find((x) => x.id === b.endAreaId);

    const startDiff = extractNumber(startA?.name) - extractNumber(startB?.name);

    if (startDiff !== 0) return startDiff;

    return extractNumber(endA?.name) - extractNumber(endB?.name);
  });

  // Remove stale routes — splice in reverse to avoid index shifting
  const validStartIds = new Set(startAreas.value.map((a) => a.id));
  const validEndIds = new Set(endAreas.value.map((a) => a.id));

  for (let i = routes.length - 1; i >= 0; i--) {
    const r = routes[i]!;
    if (!validStartIds.has(r.startAreaId) || !validEndIds.has(r.endAreaId)) {
      routes.splice(i, 1);
    }
  }

  // Auto-select first route if current selection is gone
  const stillValid = routes.find((r) => r.id === activeRouteId.value);
  if (!stillValid && routes.length > 0) {
    activeRouteId.value = routes[0]!.id;
  }
}

const isDrawingLine = computed(
  () =>
    sceneStates.value?.routeDrawing.mode === "line" &&
    sceneStates.value?.routeDrawing.activeRouteId === activeRouteId.value,
);

const isDrawingBezier = computed(
  () =>
    sceneStates.value?.routeDrawing.mode === "bezier" &&
    sceneStates.value?.routeDrawing.activeRouteId === activeRouteId.value,
);

function toggleDrawMode(mode: "line" | "bezier") {
  const rd = sceneStates.value!.routeDrawing;
  if (rd.mode === mode && rd.activeRouteId === activeRouteId.value) {
    rd.mode = "none";
    rd.activeRouteId = null;
  } else {
    rd.mode = mode;
    rd.activeRouteId = activeRouteId.value;
  }
}

function stopDrawing() {
  sceneStates.value!.routeDrawing.mode = "none";
  sceneStates.value!.routeDrawing.activeRouteId = null;
}

function removeSegment(index: number) {
  activeRoute.value?.segments.splice(index, 1);
}

function clearSegments() {
  if (activeRoute.value) activeRoute.value.segments = [];
}

// ====================== POPULATION GROUPS ======================
function addPopulationGroup() {
  sceneStates.value!.simulation.populationGroups.push({
    id: uuidv4(),
    routeId: "",
    height: 1.7,
    speed: 1.4,
    count: 10,
  });
}

function removePopulationGroup(group: PopulationGroup) {
  const index = sceneStates.value!.simulation.populationGroups.indexOf(group);
  if (index !== -1) {
    sceneStates.value!.simulation.populationGroups.splice(index, 1);
  }
}

function deleteArea(id: string) {
  sceneStates.value?.facesManagement.remove(id);
}

async function loadSimulation() {
  isLoadingSimulation.value = true;

  try {
    const workspace = await workspaceApi.getWorkspace(["simulation"]);
    const sim = workspace.simulation;

    if (!sim) return;

    // Clear existing faces of type "simulation" first
    const faces = sceneStates.value!.facesManagement;
    const existingSimFaceIds = Object.entries(faces.faces)
      .filter(([, face]) => face.type === "simulation")
      .map(([id]) => id);

    for (const id of existingSimFaceIds) {
      faces.remove(id);
    }

    for (const area of sim.areas ?? []) {
      faces.add(area.id, {
        name: area.name,
        color: area.color,
        kind: area.kind,
        type: "simulation",
        hidden: false,
        points: area.points,
        normal: new Vector3(0, 1, 0),
      });
    }

    sceneStates.value!.simulation.populationGroups.splice(
      0,
      Infinity,
      ...(sim.populationGroups ?? []),
    );
    sceneStates.value!.simulation.routes.splice(
      0,
      Infinity,
      ...(sim.routes ?? []),
    );
  } finally {
    isLoadingSimulation.value = false;
    await nextTick();
    syncRoutes();
  }
}
const activeRoute = computed(() =>
  sceneStates.value?.simulation.routes.find(
    (r) => r.id === activeRouteId.value,
  ),
);

async function saveSimulation() {
  const validRoutes = simulation.value.routes.filter(
    (r) => r.segments.length > 0,
  );
  const validRouteIds = new Set(validRoutes.map((r) => r.id));

  await workspaceApi.putSimulation({
    areas: simulation.value.areas.map((a) => ({
      id: a.id,
      name: a.name ?? "Unnamed",
      color: a.color ?? "#22ff88",
      kind: a.kind ?? "start",
      points: a.points,
    })),
    routes: validRoutes,
    populationGroups: simulation.value.populationGroups.filter(
      (g) => g.routeId !== "" && validRouteIds.has(g.routeId),
    ),
  });

  savedSimulation.value = JSON.stringify(simulation.value);
  isDirty.value = false;
}

function setMode(mode: SimulationSelectMode) {
  simulationSelectMode.value = mode;
}

function toggleMode(mode: "start" | "end") {
  simulationSelectMode.value =
    simulationSelectMode.value === mode ? "none" : mode;
}

watch(simulationSelectMode, (mode) => {
  if (!sceneStates.value) return;

  if (mode === "none") {
    sceneStates.value.selectionMode.value = "none";
    return;
  }

  sceneStates.value.selectionMode.value = "simulation";
  sceneStates.value.facesManagement.setMode("simulation");
  sceneStates.value.simulationKind.value = mode === "start" ? "start" : "end";
});

watch(
  () => simulation.value,
  () => {
    isDirty.value = JSON.stringify(simulation.value) !== savedSimulation.value;
  },
  { deep: true },
);

watch(
  () => [startAreas.value.map((a) => a.id), endAreas.value.map((a) => a.id)],
  async () => {
    await nextTick();
    syncRoutes();
  },
  { immediate: true, deep: true },
);

watch(activeRouteId, (id) => {
  if (!id && sceneStates.value?.simulation.routes.length) {
    activeRouteId.value = sceneStates.value.simulation.routes[0]!.id;
  }
});

watch(isSimulationRunning, () => {
  console.log("simulation state : ", sceneStates.value?.simulationState.value);
});

onMounted(async () => {
  await loadSimulation();
  window.addEventListener("keydown", (e) => {
    if (e.code === "Escape") stopDrawing();
  });
});
</script>

<template>
  <template v-if="currentPanel === 'simulation'">
    <div
      class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full shadow-lg"
    >
      <!-- HEADER -->
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-lg font-semibold flex items-center gap-2 text-primary">
          <Users class="h-5 w-5" />
          Crowd Simulation
        </h2>
      </div>

      <div class="space-y-3">
        <!-- Area Selection Mode -->
        <Card
          :class="[
            'transition-colors',
            simulationSelectMode === 'start' &&
              'border-blue-500 ring-2 ring-blue-200',
            simulationSelectMode === 'end' &&
              'border-purple-500 ring-2 ring-purple-200',
          ]"
        >
          <CardHeader>
            <CardTitle class="text-sm flex items-center gap-2">
              <Map class="h-4 w-4" />
              Area Selection Mode
            </CardTitle>
          </CardHeader>

          <CardContent class="space-y-2">
            <div class="flex gap-2">
              <Button
                class="flex-1"
                :class="
                  simulationSelectMode === 'start'
                    ? 'bg-blue-600 text-white hover:bg-blue-700'
                    : ''
                "
                @click="toggleMode('start')"
              >
                <LogIn class="h-4 w-4 mr-2" />
                Start
              </Button>

              <Button
                class="flex-1"
                :class="
                  simulationSelectMode === 'end'
                    ? 'bg-purple-600 text-white hover:bg-purple-700'
                    : ''
                "
                @click="toggleMode('end')"
              >
                <LogOut class="h-4 w-4 mr-2" />
                End
              </Button>

              <Button variant="outline" class="flex-1" @click="setMode('none')">
                <X class="h-4 w-4" />
              </Button>
            </div>

            <div class="text-xs text-muted-foreground">
              Create separate spawn (start) and destination (end) zones for
              crowd routing.
            </div>

            <div class="flex justify-between text-sm">
              <span class="text-muted-foreground">Defined Areas</span>
              <span class="font-medium">{{ simulationAreaCount }}</span>
            </div>
          </CardContent>
        </Card>

        <!-- Defined Areas -->
        <Card>
          <CardHeader>
            <CardTitle class="text-sm flex items-center gap-2">
              <CircleDot class="h-4 w-4" />
              Defined Areas
            </CardTitle>
          </CardHeader>

          <CardContent class="space-y-4">
            <!-- START AREAS -->
            <div>
              <div
                class="text-xs font-semibold text-blue-500 mb-2 flex items-center gap-1"
              >
                <LogIn class="h-3.5 w-3.5" />
                Start Areas
              </div>

              <div
                v-if="startAreas.length === 0"
                class="text-xs text-muted-foreground"
              >
                No start areas
              </div>

              <div v-else class="space-y-2">
                <div
                  v-for="area in startAreas"
                  :key="area.id"
                  class="rounded-md border p-3 bg-blue-500/5"
                >
                  <div class="flex justify-between items-center">
                    <div class="flex items-center gap-2">
                      <div
                        class="w-3 h-3 rounded-full border"
                        :style="{ backgroundColor: area.color ?? '#3b82f6' }"
                      />
                      <div>
                        <div class="font-medium text-sm">
                          {{ area.name ?? "Start Area" }}
                        </div>
                        <div class="text-xs text-muted-foreground">
                          {{ area.points?.length ?? 0 }} points
                        </div>
                      </div>
                    </div>

                    <Button
                      variant="ghost"
                      size="icon"
                      @click="deleteArea(area.id)"
                    >
                      <Trash2 class="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            <!-- END AREAS -->
            <div>
              <div
                class="text-xs font-semibold text-purple-500 mb-2 flex items-center gap-1"
              >
                <LogOut class="h-3.5 w-3.5" />
                End Areas
              </div>

              <div
                v-if="endAreas.length === 0"
                class="text-xs text-muted-foreground"
              >
                No end areas
              </div>

              <div v-else class="space-y-2">
                <div
                  v-for="area in endAreas"
                  :key="area.id"
                  class="rounded-md border p-3 bg-purple-500/5"
                >
                  <div class="flex justify-between items-center">
                    <div class="flex items-center gap-2">
                      <div
                        class="w-3 h-3 rounded-full border"
                        :style="{ backgroundColor: area.color ?? '#a855f7' }"
                      />
                      <div>
                        <div class="font-medium text-sm">
                          {{ area.name ?? "End Area" }}
                        </div>
                        <div class="text-xs text-muted-foreground">
                          {{ area.points?.length ?? 0 }} points
                        </div>
                      </div>
                    </div>

                    <Button
                      variant="ghost"
                      size="icon"
                      @click="deleteArea(area.id)"
                    >
                      <Trash2 class="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- ==================== ROUTES CARD (NEW) ==================== -->
        <Card>
          <CardHeader>
            <div class="flex justify-between items-center">
              <CardTitle class="text-sm flex items-center gap-2">
                <RouteIcon class="h-4 w-4" />
                Route
              </CardTitle>
            </div>
          </CardHeader>

          <CardContent class="space-y-3">
            <!-- Route Selection -->
            <div class="space-y-1">
              <label class="text-xs text-muted-foreground">
                Selected Route
              </label>

              <select v-model="activeRouteId" class="sim-select w-full">
                <option value="">Select Route</option>

                <option
                  v-for="path in simulation.routes"
                  :key="path.id"
                  :value="path.id"
                >
                  {{ path.name }}
                </option>
              </select>
            </div>

            <div v-if="activeRoute">
              <div class="border rounded-md p-3 space-y-3 bg-muted/40">
                <!-- Draw mode buttons -->
                <div class="text-xs text-muted-foreground">Draw Path</div>

                <div class="flex gap-2">
                  <Button
                    size="sm"
                    :class="
                      isDrawingLine
                        ? 'bg-blue-600 text-white hover:bg-blue-700'
                        : ''
                    "
                    variant="outline"
                    @click="toggleDrawMode('line')"
                  >
                    <Minus class="h-3.5 w-3.5 mr-1" />
                    Line
                  </Button>

                  <Button
                    size="sm"
                    :class="
                      isDrawingBezier
                        ? 'bg-orange-500 text-white hover:bg-orange-600'
                        : ''
                    "
                    variant="outline"
                    @click="toggleDrawMode('bezier')"
                  >
                    <Spline class="h-3.5 w-3.5 mr-1" />
                    Curve
                  </Button>

                  <Button
                    v-if="isDrawingLine || isDrawingBezier"
                    size="sm"
                    variant="outline"
                    @click="stopDrawing"
                  >
                    <X class="h-3.5 w-3.5" />
                  </Button>
                </div>

                <!-- Hint text while drawing -->
                <div
                  v-if="isDrawingLine || isDrawingBezier"
                  class="text-xs text-blue-400"
                >
                  Ctrl+Click on the floor to place points. Escape to stop.
                </div>

                <!-- Segment list -->
                <div v-if="activeRoute.segments.length > 0" class="space-y-1">
                  <div class="text-xs text-muted-foreground">
                    Segments ({{ activeRoute.segments.length }})
                  </div>

                  <div
                    v-for="(seg, i) in activeRoute.segments"
                    :key="i"
                    class="flex items-center justify-between text-xs bg-muted rounded px-2 py-1"
                  >
                    <span class="capitalize"> {{ seg.type }} {{ i + 1 }} </span>

                    <Button
                      size="icon"
                      variant="ghost"
                      class="h-5 w-5"
                      @click="removeSegment(i)"
                    >
                      <Trash2 class="h-3 w-3 text-destructive" />
                    </Button>
                  </div>
                </div>

                <Button
                  v-if="activeRoute.segments.length > 0"
                  size="sm"
                  variant="outline"
                  class="w-full text-destructive border-destructive"
                  @click="clearSegments"
                >
                  Clear Path
                </Button>
              </div>
            </div>

            <div v-else class="text-xs text-muted-foreground text-center py-4">
              Select a route to edit.
            </div>
          </CardContent>
        </Card>

        <!-- Population Groups -->
        <Card>
          <CardHeader>
            <div class="flex justify-between items-center">
              <CardTitle class="text-sm flex items-center gap-2">
                <UsersRound class="h-4 w-4" />
                Population Groups
              </CardTitle>

              <Button size="sm" @click="addPopulationGroup">
                <Plus class="h-4 w-4 mr-1" />
                Add
              </Button>
            </div>
          </CardHeader>

          <CardContent>
            <div
              v-if="simulation.populationGroups.length === 0"
              class="text-xs italic text-muted-foreground"
            >
              No population groups configured.
            </div>

            <div
              v-for="(group, index) in simulation.populationGroups"
              :key="group.id"
              class="border border-border rounded-md p-3 bg-muted/40 mb-3"
            >
              <div class="font-medium mb-3 flex items-center gap-2">
                <UserRound class="h-4 w-4 text-muted-foreground" />
                Group {{ index + 1 }}
              </div>

              <!-- Route Selection -->
              <div class="mb-3 flex items-center justify-between gap-4">
                <div class="flex items-center gap-2 shrink-0">
                  <RouteIcon class="h-3.5 w-3.5 text-muted-foreground" />
                  <span class="text-sm font-medium">Route</span>
                </div>

                <select
                  v-model="group.routeId"
                  class="sim-select w-full max-w-[200px]"
                >
                  <option value="">Select Route</option>
                  <option
                    v-for="r in simulation.routes"
                    :key="r.id"
                    :value="r.id"
                  >
                    {{ r.name }}
                  </option>
                </select>
              </div>

              <!-- Height -->
              <div class="sim-field">
                <span class="flex items-center gap-1.5">
                  <Ruler class="h-3.5 w-3.5 text-muted-foreground" />
                  Height
                </span>
                <Input v-model.number="group.height" type="number" step="0.1" />
              </div>

              <!-- Speed -->
              <div class="sim-field">
                <span class="flex items-center gap-1.5">
                  <Gauge class="h-3.5 w-3.5 text-muted-foreground" />
                  Speed
                </span>
                <Input v-model.number="group.speed" type="number" step="0.1" />
              </div>

              <!-- Count -->
              <div class="sim-field">
                <span class="flex items-center gap-1.5">
                  <Hash class="h-3.5 w-3.5 text-muted-foreground" />
                  Count
                </span>
                <Input v-model.number="group.count" type="number" min="1" />
              </div>

              <Button
                variant="destructive"
                size="sm"
                class="w-full mt-3"
                @click="removePopulationGroup(group)"
              >
                <Trash2 class="h-4 w-4 mr-2" />
                Remove Group
              </Button>
            </div>
          </CardContent>
        </Card>

        <!-- ACTIONS -->
        <Card>
          <CardContent class="pt-6 space-y-2">
            <Button
              class="w-full bg-emerald-500! hover:bg-emerald-600!"
              :disabled="!isDirty"
              @click="saveSimulation"
            >
              <Save class="h-4 w-4 mr-2" />
              Save Simulation
            </Button>

            <Button
              :variant="'outline'"
              class="w-full"
              :class="
                isSimulationRunning
                  ? 'border-red-500 text-red-500 hover:bg-red-50'
                  : ''
              "
              @click="runSimulation"
            >
              <Play class="h-4 w-4 mr-2" />
              {{ isSimulationRunning ? "Stop Simulation" : "Run Simulation" }}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  </template>
</template>

<style scoped lang="scss">
.sim-field {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
}

.sim-select {
  width: 100%;
  border: 1px solid hsl(var(--border));
  border-radius: 0.375rem;
  background-color: #18181b;
  color: #ffffff;
  padding: 0.35rem 0.5rem;
  font-size: 0.875rem;
  -webkit-appearance: none;
  appearance: none;
  cursor: pointer;
}

.sim-select:focus {
  outline: 2px solid hsl(var(--primary));
  border-color: transparent;
}

.sim-select option {
  background-color: #18181b;
  color: #ffffff;
}
</style>
