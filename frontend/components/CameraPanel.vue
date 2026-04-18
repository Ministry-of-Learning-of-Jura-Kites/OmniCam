<script setup lang="ts">
import { ref } from "vue";
import { Euler, Vector3, MathUtils } from "three";
import Button from "./ui/button/Button.vue";
import Input from "./ui/input/Input.vue";
import Label from "./ui/label/Label.vue";

import {
  Camera,
  Trash2,
  Eye,
  EyeOff,
  MapPinPlusInside,
  LockKeyhole,
  Dices,
  LogOut,
  Video,
  Move3D,
  Aperture,
  Palette,
} from "lucide-vue-next";
import { randomVividColor } from "~/utils/randomVividColor";
import { PANEL_KEY, SCENE_STATES_KEY } from "@/constants/state-keys";
import CameraSpawnDialog from "~/components/dialog/CameraSpawnDialog.vue";

export type Camerapreset = {
  vendor: string;
  camera: string;
  sensor_name: string;
  aspect: string;
  fov: string;
  pixel_pitch: string;
  res_w: string;
  res_h: string;
  sensor_w_mm: string;
  sensor_h_mm: string;
  focal_length: string;
  _id?: string;
};

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = inject(SCENE_STATES_KEY);

const { camPanelInfo } = inject(PANEL_KEY)!;
const { selectedCamId } = camPanelInfo;

const isCameraSpawnDialogOpen = ref(false);

type CamTab = "transform" | "lens" | "style";
const activeCamTab = ref<CamTab>("transform");

const selectedCam = computed(() =>
  selectedCamId.value ? sceneStates?.value?.cameras[selectedCamId.value] : null,
);

const isSelectedActive = computed(
  () => sceneStates?.value?.currentCamId.value === selectedCamId.value,
);

watch(
  [() => selectedCam.value?.widthRes, () => selectedCam.value?.heightRes],
  () => {
    sceneStates!.value!.aspectRatioManagement?.updateAspectFromEle();
  },
);

const moveCameraHere = (id: string) => {
  sceneStates!.value!.cameras[id]!.position = new Vector3().copy(
    sceneStates!.value!.spectatorCameraPosition,
  );
  sceneStates!.value!.cameras[id]!.rotation = new Euler().copy(
    sceneStates!.value!.spectatorCameraRotation,
  );
};

const deleteCamera = (id: string) => {
  if (sceneStates!.value!.currentCamId.value == id) {
    sceneStates!.value!.currentCamId.value = null;
  }
  delete sceneStates!.value!.cameras[id];
};

function randomNewFrustumColor() {
  const cam = sceneStates!.value!.cameras[selectedCamId.value!]!;
  const color = randomVividColor();
  cam.frustumColor.r = color.r;
  cam.frustumColor.g = color.g;
  cam.frustumColor.b = color.b;
}

function onToggleLockPosition() {
  const cam = sceneStates!.value!.cameras[selectedCamId.value!]!;
  cam.isHidingArrows = cam.isLockingPosition;
}

function onToggleLockRotation() {
  const cam = sceneStates!.value!.cameras[selectedCamId.value!]!;
  cam.isHidingWheels = cam.isLockingRotation;
}

function onFovChange() {
  const cam = sceneStates!.value!.cameras[selectedCamId.value!]!;
  if (cam.fov > 179) {
    cam.isHidingFrustum = true;
  }
}

function getUniqueCameraName(baseName: string) {
  const names = Object.values(sceneStates!.value!.cameras).map((c) => c.name);
  if (!names.includes(baseName)) return baseName;
  let i = 2;
  while (names.includes(`${baseName} (${i})`)) {
    i++;
  }
  return `${baseName} (${i})`;
}

const isLockingRotation = computed({
  get() {
    return (
      sceneStates?.value?.cameras[selectedCamId.value!]?.isLockingRotation ||
      props.workspace != "me"
    );
  },
  set(val: boolean) {
    if (
      !selectedCamId.value ||
      !sceneStates?.value?.cameras[selectedCamId.value]
    )
      return;
    sceneStates.value.cameras[selectedCamId.value]!.isLockingRotation = val;
  },
});

const isLockingPosition = computed({
  get() {
    return (
      sceneStates?.value?.cameras[selectedCamId.value!]?.isLockingPosition ||
      props.workspace != "me"
    );
  },
  set(val: boolean) {
    if (
      !selectedCamId.value ||
      !sceneStates?.value?.cameras[selectedCamId.value]
    )
      return;
    sceneStates.value.cameras[selectedCamId.value]!.isLockingPosition = val;
  },
});

const handleSpawnCamera = (preset: Camerapreset) => {
  if (!sceneStates) return;
  const camId = sceneStates.value!.cameraManagement.spawnCameraHere();
  const cam = sceneStates.value!.cameras[camId];
  if (cam) {
    cam.name = getUniqueCameraName(`${preset.vendor} ${preset.camera}`);
    cam.fov = Number(preset.fov);
    cam.widthRes = Number(preset.res_w);
    cam.heightRes = Number(preset.res_h);
    cam.frustumLength = Number(preset.focal_length) * 50;
    sceneStates.value!.markedForCheck.value = true;
    if (selectedCamId.value !== undefined) {
      selectedCamId.value = camId;
    }
  }
};

const createRotationRef = (axis: "x" | "y" | "z") => {
  return computed({
    get() {
      if (!selectedCam.value) return 0;
      return (
        Math.round(MathUtils.radToDeg(selectedCam.value.rotation[axis]) * 100) /
        100
      );
    },
    set(value: number) {
      if (!selectedCam.value) return;
      selectedCam.value.rotation[axis] = MathUtils.degToRad(value);
      sceneStates!.value!.markedForCheck.value = true;
    },
  });
};

const angleX = createRotationRef("x");
const angleY = createRotationRef("y");
const angleZ = createRotationRef("z");

function signedAngle(a: Vector3, b: Vector3, normal: Vector3) {
  const angle = a.angleTo(b);
  const sign = Math.sign(a.clone().cross(b).dot(normal)) || -1;
  return angle * sign;
}

const directionAngles = computed(() => {
  if (!selectedCam.value) return { x: 0, y: 0, z: 0 };
  const forward = new Vector3(0, 0, 1)
    .applyEuler(selectedCam.value.rotation)
    .normalize();
  return {
    x:
      Math.round(
        MathUtils.radToDeg(
          signedAngle(forward, new Vector3(1, 0, 0), new Vector3(0, 1, 0)),
        ) * 100,
      ) / 100,
    y:
      Math.round(
        MathUtils.radToDeg(
          signedAngle(forward, new Vector3(0, 1, 0), new Vector3(0, 0, 1)),
        ) * 100,
      ) / 100,
    z:
      Math.round(
        MathUtils.radToDeg(
          signedAngle(forward, new Vector3(0, 0, 1), new Vector3(0, -1, 0)),
        ) * 100,
      ) / 100,
  };
});

function toggleCamera(camId: string) {
  if (isSelectedActive.value) {
    sceneStates!.value!.cameraManagement.switchToSpectator();
  } else {
    sceneStates!.value!.cameraManagement.switchToCam(camId);
  }
}
</script>

<template>
  <TooltipProvider>
    <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full">
      <!-- Header -->
      <div class="flex items-center justify-between mb-5">
        <h2 class="text-lg font-semibold flex items-center gap-2">
          <Camera class="h-5 w-5" />
          Camera Gallery
        </h2>
        <Button
          size="sm"
          :disabled="props.workspace != 'me'"
          @click="
            isCameraSpawnDialogOpen = true;
            $event.currentTarget.blur();
          "
        >
          <MapPinPlusInside class="h-4 w-4" />
        </Button>
      </div>

      <!-- Camera Dropdown -->
      <div class="mb-4">
        <Label for="camera-select" class="mb-1.5 block text-sm"
          >Select Camera</Label
        >
        <select
          id="camera-select"
          v-model="selectedCamId"
          class="w-full border border-border rounded-md px-3 py-2 bg-background text-foreground text-sm focus:outline-none focus:ring-1 focus:ring-ring"
        >
          <option
            v-for="[camId, camera] of Object.entries(
              sceneStates?.cameras ?? {},
            )"
            :key="camId"
            :value="camId"
          >
            {{ camera.name }} (VFOV: {{ camera.fov }}°)
          </option>
        </select>
      </div>

      <!-- Exit Camera -->
      <div class="mb-5">
        <ClientOnly>
          <Button
            size="sm"
            :variant="isSelectedActive ? 'secondary' : 'outline'"
            :class="
              isSelectedActive
                ? 'bg-red-500 hover:bg-red-700 text-white w-full'
                : 'w-full'
            "
            @click="toggleCamera(selectedCamId!)"
          >
            <template v-if="isSelectedActive">
              <LogOut class="h-4 w-4 mr-2" />
              <!-- TODO: If select a different camera while still within the camera's view, a warning to exit should still appear. -->
              Exit Camera
            </template>
            <template v-else>
              <Video class="h-4 w-4 mr-2" />
              View Camera
            </template>
          </Button>
        </ClientOnly>
      </div>

      <hr class="border-border mb-3" />

      <!-- Camera Properties-->
      <div
        v-if="selectedCamId && sceneStates?.cameras[selectedCamId]"
        class="space-y-4"
      >
        <!-- Camera Name -->
        <div>
          <Label for="camera-name" class="mb-1.5 block text-sm">Name</Label>
          <Input
            id="camera-name"
            v-model="sceneStates.cameras[selectedCamId]!.name"
            :disabled="props.workspace != 'me'"
            disabled-class="disabled-input"
          />
        </div>

        <!-- Sub-tabs -->
        <div
          class="flex rounded-md border border-border overflow-hidden text-sm font-medium"
        >
          <button
            class="flex-1 flex items-center justify-center gap-2 py-2 transition-colors"
            :class="
              activeCamTab === 'transform'
                ? 'bg-primary text-primary-foreground'
                : 'bg-background text-muted-foreground hover:bg-muted'
            "
            @click="activeCamTab = 'transform'"
          >
            <Move3D class="h-4 w-4" />
            Transform
          </button>
          <button
            class="flex-1 flex items-center justify-center gap-2 py-2 border-x border-border transition-colors"
            :class="
              activeCamTab === 'lens'
                ? 'bg-primary text-primary-foreground'
                : 'bg-background text-muted-foreground hover:bg-muted'
            "
            @click="activeCamTab = 'lens'"
          >
            <Aperture class="h-4 w-4" />
            Lens
          </button>
          <button
            class="flex-1 flex items-center justify-center gap-2 py-2 transition-colors"
            :class="
              activeCamTab === 'style'
                ? 'bg-primary text-primary-foreground'
                : 'bg-background text-muted-foreground hover:bg-muted'
            "
            @click="activeCamTab = 'style'"
          >
            <Palette class="h-4 w-4" />
            Style
          </button>
        </div>

        <!-- ══ TRANSFORM TAB ══ -->
        <div v-if="activeCamTab === 'transform'" class="space-y-5">
          <!-- Position -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <Label class="text-sm font-semibold">Position</Label>
              <label
                class="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer select-none"
              >
                <input
                  id="lock-position"
                  v-model="isLockingPosition"
                  :disabled="props.workspace != 'me'"
                  type="checkbox"
                  @change="onToggleLockPosition"
                />
                <LockKeyhole class="h-3.5 w-3.5" />
                Lock
              </label>
            </div>
            <div class="grid grid-cols-3 gap-2">
              <div>
                <span class="axis-label axis-x">X</span>
                <Input
                  id="pos-x"
                  v-model.number="
                    sceneStates.cameras[selectedCamId]!.position.x
                  "
                  :disabled="isLockingPosition"
                  disabled-class="disabled-input"
                  type="number"
                />
              </div>
              <div>
                <span class="axis-label axis-y">Y</span>
                <Input
                  id="pos-y"
                  v-model.number="
                    sceneStates.cameras[selectedCamId]!.position.y
                  "
                  :disabled="isLockingPosition"
                  disabled-class="disabled-input"
                  type="number"
                />
              </div>
              <div>
                <span class="axis-label axis-z">Z</span>
                <Input
                  id="pos-z"
                  v-model.number="
                    sceneStates.cameras[selectedCamId]!.position.z
                  "
                  :disabled="isLockingPosition"
                  disabled-class="disabled-input"
                  type="number"
                />
              </div>
            </div>
          </div>

          <hr class="border-border" />

          <!-- Rotation -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <Label class="text-sm font-semibold">Rotation</Label>
              <label
                class="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer select-none"
              >
                <input
                  id="lock-rotation"
                  v-model="
                    sceneStates.cameras[selectedCamId]!.isLockingRotation
                  "
                  :disabled="props.workspace != 'me'"
                  type="checkbox"
                  @change="onToggleLockRotation"
                />
                <LockKeyhole class="h-3.5 w-3.5" />
                Lock
              </label>
            </div>
            <div class="grid grid-cols-3 gap-2">
              <div>
                <span class="axis-label axis-x">RX</span>
                <Input
                  id="angle-x"
                  v-model.number="angleX"
                  :disabled="isLockingRotation"
                  disabled-class="disabled-input"
                  type="number"
                  step="0.1"
                />
              </div>
              <div>
                <span class="axis-label axis-y">RY</span>
                <Input
                  id="angle-y"
                  v-model.number="angleY"
                  :disabled="isLockingRotation"
                  disabled-class="disabled-input"
                  type="number"
                  step="0.1"
                />
              </div>
              <div>
                <span class="axis-label axis-z">RZ</span>
                <Input
                  id="angle-z"
                  v-model.number="angleZ"
                  :disabled="isLockingRotation"
                  disabled-class="disabled-input"
                  type="number"
                  step="0.1"
                />
              </div>
            </div>
          </div>

          <!-- Angles Vector -->
          <div v-if="selectedCam" class="space-y-2">
            <Label class="text-xs text-muted-foreground uppercase tracking-wide"
              >Angles Vector</Label
            >
            <div
              class="grid grid-cols-3 gap-2 rounded-md bg-muted/40 px-3 py-2.5"
            >
              <div class="flex flex-col gap-0.5">
                <span class="axis-label axis-x">X-Axis</span>
                <span class="font-mono text-sm">{{ directionAngles.x }}°</span>
              </div>
              <div class="flex flex-col gap-0.5">
                <span class="axis-label axis-y">Y-Axis</span>
                <span class="font-mono text-sm">{{ directionAngles.y }}°</span>
              </div>
              <div class="flex flex-col gap-0.5">
                <span class="axis-label axis-z">Z-Axis</span>
                <span class="font-mono text-sm">{{ directionAngles.z }}°</span>
              </div>
            </div>
          </div>

          <hr class="border-border" />

          <!-- Gizmo toggles -->
          <div class="space-y-2">
            <div class="grid grid-cols-2 gap-2">
              <Button
                size="sm"
                variant="outline"
                :disabled="isLockingPosition"
                disabled-class="disabled-input"
                @click="
                  sceneStates.cameras[selectedCamId]!.isHidingArrows =
                    !sceneStates.cameras[selectedCamId]!.isHidingArrows
                "
              >
                <template
                  v-if="sceneStates.cameras[selectedCamId]!.isLockingPosition"
                >
                  <LockKeyhole class="h-4 w-4 mr-1.5" />
                </template>
                <template v-else>
                  <Eye
                    v-if="!sceneStates.cameras[selectedCamId]!.isHidingArrows"
                    class="h-4 w-4 mr-1.5"
                  />
                  <EyeOff v-else class="h-4 w-4 mr-1.5" />
                </template>
                Arrows
              </Button>
              <Button
                size="sm"
                variant="outline"
                :disabled="
                  sceneStates.cameras[selectedCamId]!.isLockingRotation ||
                  props.workspace != 'me'
                "
                disabled-class="disabled-input"
                @click="
                  sceneStates.cameras[selectedCamId]!.isHidingWheels =
                    !sceneStates.cameras[selectedCamId]!.isHidingWheels
                "
              >
                <template
                  v-if="sceneStates.cameras[selectedCamId]!.isLockingRotation"
                >
                  <LockKeyhole class="h-4 w-4 mr-1.5" />
                </template>
                <template v-else>
                  <Eye
                    v-if="!sceneStates.cameras[selectedCamId]!.isHidingWheels"
                    class="h-4 w-4 mr-1.5"
                  />
                  <EyeOff v-else class="h-4 w-4 mr-1.5" />
                </template>
                Wheels
              </Button>
            </div>

            <!-- Teleport / Delete -->
            <div class="space-y-2">
              <Button
                size="sm"
                variant="outline"
                class="w-full"
                :disabled="
                  sceneStates.cameras[selectedCamId]!.isLockingPosition ||
                  sceneStates.cameras[selectedCamId]!.isLockingRotation ||
                  props.workspace != 'me'
                "
                disabled-class="disabled-input"
                @click="moveCameraHere(selectedCamId!)"
              >
                Teleport Camera to Me
              </Button>

              <Button
                size="sm"
                variant="outline"
                class="w-full text-destructive hover:bg-destructive/10"
                :disabled="
                  sceneStates.currentCamId.value == selectedCamId ||
                  props.workspace != 'me'
                "
                @click="deleteCamera(selectedCamId)"
              >
                <Trash2 class="h-4 w-4 mr-2" />
                Delete Camera
              </Button>
            </div>
          </div>
        </div>

        <!-- ══ LENS TAB ══ -->
        <div v-if="activeCamTab === 'lens'" class="space-y-6">
          <!-- Section: Lens Configuration -->
          <section
            class="space-y-4 p-4 rounded-xl bg-muted/30 border border-border/50"
          >
            <div class="flex items-center gap-2 mb-2">
              <Label
                class="text-sm font-bold uppercase tracking-wider text-muted-foreground"
                >Lens Configuration</Label
              >
            </div>

            <!-- Resolution / Aspect Ratio -->
            <div class="space-y-2">
              <Label class="text-[11px] text-muted-foreground uppercase"
                >Resolution (W : H)</Label
              >
              <div class="flex items-center gap-2">
                <div class="relative flex-1">
                  <Input
                    v-model.number="
                      sceneStates.cameras[selectedCamId]!.widthRes
                    "
                    :disabled="props.workspace != 'me'"
                    type="number"
                    class="pl-8"
                  />
                  <span
                    class="absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold opacity-40"
                    >W</span
                  >
                </div>

                <span class="text-muted-foreground font-light">:</span>

                <div class="relative flex-1">
                  <Input
                    v-model.number="
                      sceneStates.cameras[selectedCamId]!.heightRes
                    "
                    :disabled="props.workspace != 'me'"
                    type="number"
                    class="pl-8"
                  />
                  <span
                    class="absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold opacity-40"
                    >H</span
                  >
                </div>
              </div>
            </div>

            <hr class="border-border/40" />

            <!-- FOV Settings -->
            <div class="space-y-2">
              <div class="flex justify-between items-center">
                <Label
                  for="fov"
                  class="text-[11px] text-muted-foreground uppercase"
                  >Vertical FOV</Label
                >
                <span class="text-[11px] font-mono text-primary"
                  >{{ sceneStates.cameras[selectedCamId]!.fov }}°</span
                >
              </div>
              <Input
                id="fov"
                v-model.number="sceneStates.cameras[selectedCamId]!.fov"
                :disabled="props.workspace != 'me'"
                type="number"
                min="10"
                max="180"
                @change="onFovChange()"
              />
            </div>

            <hr class="border-border/40" />

            <!-- Distortion Controls -->
            <div class="space-y-3">
              <Label
                class="text-[11px] text-muted-foreground uppercase block mb-1"
                >Optical Distortion</Label
              >

              <div class="grid grid-cols-2 gap-4">
                <label
                  class="flex items-center gap-2.5 p-2 rounded-lg border border-border/50 bg-background/50 cursor-pointer hover:bg-accent transition-colors select-none"
                  :class="{
                    'opacity-50 cursor-not-allowed': props.workspace != 'me',
                  }"
                >
                  <input
                    v-model="
                      sceneStates.cameras[selectedCamId]!.distortion.enabled
                    "
                    type="checkbox"
                    :disabled="props.workspace != 'me'"
                    class="accent-primary h-4 w-4"
                  />
                  <span class="text-xs font-medium">Enable</span>
                </label>

                <label
                  class="flex items-center gap-2.5 p-2 rounded-lg border border-border/50 bg-background/50 cursor-pointer hover:bg-accent transition-colors select-none"
                  :class="{
                    'opacity-50 cursor-not-allowed':
                      props.workspace != 'me' ||
                      !sceneStates.cameras[selectedCamId]!.distortion.enabled,
                  }"
                >
                  <input
                    v-model="
                      sceneStates.cameras[selectedCamId]!.distortion.isFisheye
                    "
                    type="checkbox"
                    :disabled="
                      props.workspace != 'me' ||
                      !sceneStates.cameras[selectedCamId]!.distortion.enabled
                    "
                    class="accent-primary h-4 w-4"
                  />
                  <span class="text-xs font-medium">Fisheye</span>
                </label>
              </div>
            </div>
          </section>
        </div>

        <!-- ══ STYLE TAB ══ -->
        <div v-if="activeCamTab === 'style'" class="space-y-3">
          <!-- Section: Frustum Style -->
          <section
            class="space-y-4 p-4 rounded-xl bg-muted/30 border border-border/50"
          >
            <div class="flex items-center justify-between">
              <Label
                class="text-sm font-bold uppercase tracking-wider text-muted-foreground"
                >Frustum Style</Label
              >
              <div class="flex gap-2">
                <!-- Visibility Toggle -->
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      size="icon"
                      variant="ghost"
                      :disabled="sceneStates?.cameras[selectedCamId]!.fov > 179"
                      @click="
                        sceneStates!.cameras[selectedCamId]!.isHidingFrustum =
                          !sceneStates!.cameras[selectedCamId]!.isHidingFrustum
                      "
                    >
                      <Eye
                        v-if="
                          !sceneStates?.cameras[selectedCamId]!.isHidingFrustum
                        "
                        class="h-4 w-4"
                      />
                      <EyeOff v-else class="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent
                    v-if="sceneStates?.cameras[selectedCamId]!.fov > 179"
                  >
                    FOV is too high for visualization
                  </TooltipContent>
                </Tooltip>

                <!-- Random Color -->
                <Button
                  size="icon"
                  variant="ghost"
                  @click="randomNewFrustumColor()"
                >
                  <Dices class="h-4 w-4" />
                </Button>
              </div>
            </div>

            <!-- Color RGB Grid -->
            <div class="space-y-3">
              <div class="flex items-center gap-2">
                <div
                  class="h-2 w-2 rounded-full"
                  :style="{
                    backgroundColor: `rgb(${sceneStates.cameras[selectedCamId]!.frustumColor.r * 255}, ${sceneStates.cameras[selectedCamId]!.frustumColor.g * 255}, ${sceneStates.cameras[selectedCamId]!.frustumColor.b * 255})`,
                  }"
                ></div>
                <span class="text-xs font-medium">Color Channels (0-1)</span>
              </div>

              <div class="grid grid-cols-3 gap-3">
                <div class="relative">
                  <Input
                    v-model.number="
                      sceneStates.cameras[selectedCamId]!.frustumColor.r
                    "
                    type="number"
                    step="0.1"
                    min="0"
                    max="1"
                    class="pl-7"
                  />
                  <span
                    class="absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold text-red-500"
                    >R</span
                  >
                </div>
                <div class="relative">
                  <Input
                    v-model.number="
                      sceneStates.cameras[selectedCamId]!.frustumColor.g
                    "
                    type="number"
                    step="0.1"
                    min="0"
                    max="1"
                    class="pl-7"
                  />
                  <span
                    class="absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold text-green-500"
                    >G</span
                  >
                </div>
                <div class="relative">
                  <Input
                    v-model.number="
                      sceneStates.cameras[selectedCamId]!.frustumColor.b
                    "
                    type="number"
                    step="0.1"
                    min="0"
                    max="1"
                    class="pl-7"
                  />
                  <span
                    class="absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold text-blue-500"
                    >B</span
                  >
                </div>
              </div>
            </div>

            <!-- Display Settings -->
            <div class="grid grid-cols-2 gap-4 pt-2">
              <div class="space-y-2">
                <Label class="text-[11px] text-muted-foreground uppercase"
                  >Opacity</Label
                >
                <Input
                  v-model.number="
                    sceneStates.cameras[selectedCamId]!.frustumColor.a
                  "
                  type="number"
                  step="0.1"
                  min="0"
                  max="1"
                />
              </div>
              <div class="space-y-2">
                <Label class="text-[11px] text-muted-foreground uppercase"
                  >Length</Label
                >
                <Input
                  v-model.number="
                    sceneStates.cameras[selectedCamId]!.frustumLength
                  "
                  type="number"
                  min="0"
                />
              </div>
            </div>
          </section>

          <!-- Section: Camera Body Style -->
          <section
            class="space-y-4 p-4 rounded-xl bg-muted/30 border border-border/50"
          >
            <Label
              class="text-sm font-bold uppercase tracking-wider text-muted-foreground"
              >Camera Body</Label
            >
            <div class="text-xs text-muted-foreground italic">
              Additional camera styles will go here...
            </div>
          </section>
        </div>
      </div>
      <!-- end camera properties -->

      <CameraSpawnDialog
        v-model="isCameraSpawnDialogOpen"
        :on-confirm="handleSpawnCamera"
      />
    </div>
  </TooltipProvider>
</template>

<style lang="scss" scoped>
input {
  field-sizing: content;
}

input[type="number"]::-webkit-outer-spin-button,
input[type="number"]::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

input[type="number"] {
  appearance: textfield;
  -moz-appearance: textfield;
}

input[type="number"] {
  border-radius: 5px;
  border: 1px solid black;
  outline: 1px solid white;
  box-sizing: border-box;
  margin-top: 4px;
}

.disabled-input {
  background-color: var(--background);
  cursor: not-allowed;
  opacity: 0.5;
}

.axis-label {
  display: block;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  margin-bottom: 2px;
}
.axis-x {
  color: #22c55e;
}
.axis-y {
  color: #ef4444;
}
.axis-z {
  color: #3b82f6;
}

Button {
  transition-property:
    color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}
</style>
