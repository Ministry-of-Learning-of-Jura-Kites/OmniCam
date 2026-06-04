<script setup lang="ts">
import { computed, inject } from "vue";
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
} from "lucide-vue-next";

const sceneStates = inject(SCENE_STATES_KEY)!;
const { currentPanel } = inject(PANEL_KEY)!;

const route = useRoute();
const runtimeConfig = useRuntimeConfig();

const projectId = route.params.projectId as string;
const modelId = route.params.modelId as string;
const workspaceId = route.params.workspaceId as string;

const savedSimulation = ref("");
const isDirty = ref<boolean>(false);

const workspaceApi = useWorkspaceApi(
  projectId,
  modelId,
  workspaceId,
  runtimeConfig,
);

const populationGroups = computed({
  get: () => simulation.value.populationGroups,
  set: (v) => (simulation.value.populationGroups = v),
});

const simulation = computed(() => {
  const base = sceneStates.value?.simulation;

  if (!base) {
    return {
      areas: [],
      populationGroups: [],
    };
  }

  const areas = Object.entries(sceneStates.value!.facesManagement.faces)
    .filter(([, face]) => face.type === "simulation")
    .map(([id, face]) => ({
      id,
      name: face.name,
      color: face.color,
      kind: face.kind,
      points: face.points,
    }));

  return {
    ...base,
    areas,
  };
});

const simulationAreaCount = computed(() => simulation.value.areas.length);

watch(simulationAreaCount, (newCount) => {
  console.log("Simulation area count changed to:", newCount);
});

const startAreas = computed(() =>
  simulation.value.areas.filter((a) => a.kind === "start"),
);

const endAreas = computed(() =>
  simulation.value.areas.filter((a) => a.kind === "end"),
);
watch(startAreas, (newStartAreas) => {
  console.log("Start areas updated:", newStartAreas);
});

watch(endAreas, (newEndAreas) => {
  console.log("End areas updated:", newEndAreas);
});

type SimulationSelectMode = "none" | "start" | "end";

const simulationSelectMode = ref<SimulationSelectMode>("none");

async function loadSimulation() {
  const workspace = await workspaceApi.getWorkspace(["simulation"]);

  sceneStates.value!.simulation.populationGroups =
    workspace.simulation?.populationGroups ?? [];

  const areas = workspace.simulation?.areas ?? [];

  for (const area of areas) {
    sceneStates.value!.facesManagement.add(area.id, {
      name: area.name,
      color: area.color ?? "#22ff88",
      kind: area.kind,
      type: "simulation",
      hidden: false,
      points: area.points,
      normal: new Vector3(0, 1, 0),
    });
  }

  savedSimulation.value = JSON.stringify(workspace.simulation ?? {});
  isDirty.value = false;
}
function setMode(mode: SimulationSelectMode) {
  simulationSelectMode.value = mode;
}

function toggleMode(mode: "start" | "end") {
  simulationSelectMode.value =
    simulationSelectMode.value === mode ? "none" : mode;
}

watch(simulationSelectMode, (newMode) => {
  console.log("Simulation select mode changed to:", newMode);
});

function addPopulationGroup() {
  populationGroups.value.push({
    id: uuidv4(),
    startAreaId: "",
    endAreaId: "",
    height: 1.7,
    speed: 1.4,
    count: 10,
  });
}

function removePopulationGroup(group: PopulationGroup) {
  const index = populationGroups.value.indexOf(group);
  if (index !== -1) populationGroups.value.splice(index, 1);
}

function deleteArea(id: string) {
  sceneStates.value?.facesManagement.remove(id);

  // sceneStates.value!.simulation.areas =
  //   sceneStates.value!.simulation.areas.filter((area) => area.id !== id);
}
async function saveSimulation() {
  await workspaceApi.putSimulation(sceneStates.value!.simulation);
  isDirty.value = false;
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
onMounted(async () => {
  await loadSimulation();
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

        <Card>
          <CardHeader>
            <CardTitle class="text-sm flex items-center gap-2">
              <CircleDot class="h-4 w-4" />
              Defined Areas
            </CardTitle>
          </CardHeader>

          <CardContent class="space-y-4">
            <!-- START -->
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

            <!-- END -->
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
              v-if="populationGroups.length === 0"
              class="text-xs italic text-muted-foreground"
            >
              No population groups configured.
            </div>

            <div
              v-for="(group, index) in populationGroups"
              :key="group.id"
              class="border border-border rounded-md p-3 bg-muted/40 mb-3"
            >
              <div class="font-medium mb-3 flex items-center gap-2">
                <UserRound class="h-4 w-4 text-muted-foreground" />
                Group {{ index + 1 }}
              </div>

              <div class="flex justify-between items-center gap-3 mb-3 text-sm">
                <span class="flex items-center gap-1.5">
                  <LogIn class="h-3.5 w-3.5 text-muted-foreground" />
                  Start Area
                </span>
                <select
                  v-model="group.startAreaId"
                  class="w-full border border-border rounded-md px-3 py-2 bg-background text-white text-sm focus:outline-none focus:ring-1 focus:ring-ring"
                >
                  <option value="">Select Area</option>
                  <option
                    v-for="area in startAreas"
                    :key="area.id"
                    :value="area.id"
                  >
                    {{ area.name ?? area.id }}
                  </option>
                </select>
              </div>

              <div class="flex justify-between items-center gap-3 mb-3 text-sm">
                <span class="flex items-center gap-1.5">
                  <LogOut class="h-3.5 w-3.5 text-muted-foreground" />
                  End Area
                </span>
                <select
                  v-model="group.endAreaId"
                  class="w-full border border-border rounded-md px-3 py-2 bg-background text-white text-sm focus:outline-none focus:ring-1 focus:ring-ring"
                >
                  <option value="">Select Area</option>
                  <option
                    v-for="area in endAreas"
                    :key="area.id"
                    :value="area.id"
                  >
                    {{ area.name ?? area.id }}
                  </option>
                </select>
              </div>

              <div class="sim-field">
                <span class="flex items-center gap-1.5">
                  <Ruler class="h-3.5 w-3.5 text-muted-foreground" />
                  Height
                </span>
                <Input v-model.number="group.height" type="number" step="0.1" />
              </div>

              <div class="sim-field">
                <span class="flex items-center gap-1.5">
                  <Gauge class="h-3.5 w-3.5 text-muted-foreground" />
                  Speed
                </span>
                <Input v-model.number="group.speed" type="number" step="0.1" />
              </div>

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
                class="w-full mt-2"
                @click="removePopulationGroup(group)"
              >
                <Trash2 class="h-4 w-4 mr-2" />
                Remove Group
              </Button>
            </div>
          </CardContent>
        </Card>

        <!-- ===================================================== -->
        <!-- ACTIONS -->
        <!-- ===================================================== -->

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

            <Button variant="outline" class="w-full">
              <Play class="h-4 w-4 mr-2" />
              Run Simulation
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
  width: 140px;
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
