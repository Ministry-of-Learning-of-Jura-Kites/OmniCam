<script setup lang="ts">
import { ref } from "vue";
import * as THREE from "three";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import Button from "./ui/button/Button.vue";
import Input from "./ui/input/Input.vue";
import Label from "./ui/label/Label.vue";
import {
  Camera,
  Trash2,
  Eye,
  EyeOff,
  Settings,
  ChevronLeft,
  ChevronDown,
  MapPinPlusInside,
  LockKeyhole,
  Pyramid,
  Dices,
} from "lucide-vue-next";
import { SCENE_STATES_KEY } from "./3d/scene-states-provider/create-scene-states";
import { randomVividColor } from "~/utils/randomVividColor";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const selectedCamId = ref<string | null>(null);

const isCameraPropertiesOpen = ref(true);
const isFrustumPropertiesOpen = ref(true);

const selectedCam = computed(() =>
  selectedCamId.value ? sceneStates.cameras[selectedCamId.value] : null,
);

watch(
  [() => selectedCam.value?.aspectWidth, () => selectedCam.value?.aspectHeight],
  () => {
    sceneStates.aspectRatioManagement?.updateAspectFromEle();
  },
);

const spawnCamera = () => {
  const newCamId = sceneStates.cameraManagement.spawnCameraHere();
  if (newCamId) {
    selectedCamId.value = newCamId;
  }
};

const moveCameraHere = (id: string) => {
  sceneStates.cameras[id]!.position = new THREE.Vector3().copy(
    sceneStates.spectatorCameraPosition,
  );
  sceneStates.cameras[id]!.rotation = new THREE.Euler().copy(
    sceneStates.spectatorCameraRotation,
  );
};

const deleteCamera = (id: string) => {
  if (sceneStates.currentCamId.value == id) {
    sceneStates.currentCamId.value = null;
  }
  delete sceneStates.cameras[id];
};

function randomNewFrustumColor() {
  const cam = sceneStates.cameras[selectedCamId.value!]!;
  const color = randomVividColor();
  cam.frustumColor.r = color.r;
  cam.frustumColor.g = color.g;
  cam.frustumColor.b = color.b;
}

function onToggleLockPosition() {
  const cam = sceneStates.cameras[selectedCamId.value!]!;

  if (cam.isLockingPosition) {
    cam.isHidingArrows = true;
  } else {
    cam.isHidingArrows = false;
  }
}

function onToggleLockRotation() {
  const cam = sceneStates.cameras[selectedCamId.value!]!;

  if (cam.isLockingRotation) {
    cam.isHidingWheels = true;
  } else {
    cam.isHidingWheels = false;
  }
}

const isLockingRotation = computed(() => {
  return (
    sceneStates.currentCam.value.isLockingRotation || props.workspace == null
  );
});
const isLockingPosition = computed(() => {
  return (
    sceneStates.currentCam.value.isLockingPosition || props.workspace == null
  );
});
</script>

<template>
  <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto h-full">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        <Camera class="h-5 w-5" />
        Camera Gallery
      </h2>
      <!-- <Button
        size="sm"
        @click="sceneStates.cameraManagement.switchToSpectator()"
      >
        <RotateCcw class="h-4 w-4" />
      </Button> -->
      <Button
        size="sm"
        :disabled="props.workspace == null"
        @click="
          spawnCamera();
          $event.currentTarget.blur();
        "
      >
        <MapPinPlusInside class="h-4 w-4" />
      </Button>
    </div>

    <!-- Camera Dropdown -->
    <div class="mb-3">
      <Label for="camera-select" class="mb-1 block">Select Camera</Label>
      <select
        id="camera-select"
        v-model="selectedCamId"
        class="w-full border rounded px-3 py-2 bg-background text-foreground"
      >
        <option
          v-for="[camId, camera] of Object.entries(sceneStates.cameras)"
          :key="camId"
          :value="camId"
        >
          {{ camera.name }} (VFOV: {{ camera.fov }}°)
        </option>
      </select>
      <div class="flex gap-2 mt-2"></div>
    </div>

    <div class="space-y-3">
      <!-- Camera Properties -->
      <Card v-if="selectedCamId && sceneStates.cameras[selectedCamId]">
        <CardHeader
          class="cursor-pointer flex items-center justify-between"
          @click="
            isCameraPropertiesOpen = !isCameraPropertiesOpen;
            sceneStates.markedForCheck.add(selectedCamId);
          "
        >
          <CardTitle class="text-base flex items-center gap-2">
            <Settings class="h-4 w-4" />
            Camera Properties
          </CardTitle>
          <span class="text-sm">
            <ChevronDown v-if="isCameraPropertiesOpen" class="inline h-4 w-4" />
            <ChevronLeft v-else class="inline h-4 w-4"
          /></span>
        </CardHeader>
        <CardContent v-if="isCameraPropertiesOpen" class="space-y-2">
          <div>
            <Label for="camera-name">Name</Label>
            <Input
              id="camera-name"
              v-model="sceneStates.cameras[selectedCamId]!.name"
              :disabled="props.workspace == null"
              disabled-class="disabled-input"
              @change="sceneStates.markedForCheck.add(selectedCamId)"
            />
          </div>

          <div class="grid grid-cols-3 gap-2">
            <div>
              <Label for="pos-x">X</Label>
              <Input
                id="pos-x"
                v-model.number="sceneStates.cameras[selectedCamId]!.position.x"
                :disabled="isLockingRotation"
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
            <div>
              <Label for="pos-y">Y</Label>
              <Input
                id="pos-y"
                v-model.number="sceneStates.cameras[selectedCamId]!.position.y"
                :disabled="isLockingRotation"
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
            <div>
              <Label for="pos-z">Z</Label>
              <Input
                id="pos-z"
                v-model.number="sceneStates.cameras[selectedCamId]!.position.z"
                :disabled="isLockingRotation"
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
          </div>

          <div class="flex items-center gap-2">
            <input
              id="lock-position"
              v-model="isLockingPosition"
              :disabled="props.workspace == null"
              type="checkbox"
              @change="onToggleLockPosition"
            />
            <label for="lock-position">Lock Position</label>
          </div>

          <div class="grid grid-cols-3 gap-2">
            <div>
              <Label for="angle-x"
                ><p>θ<sub>x</sub></p></Label
              >
              <Input
                id="angle-x"
                v-model.number="sceneStates.cameras[selectedCamId]!.rotation.x"
                :disabled="
                  sceneStates.cameras[selectedCamId]!.isLockingRotation ||
                  props.workspace == null
                "
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
            <div>
              <Label for="angle-y"
                ><p>θ<sub>y</sub></p></Label
              >
              <Input
                id="angle-y"
                v-model.number="sceneStates.cameras[selectedCamId]!.rotation.y"
                :disabled="
                  sceneStates.cameras[selectedCamId]!.isLockingRotation ||
                  props.workspace == null
                "
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
            <div>
              <Label for="angle-z"
                ><p>θ<sub>z</sub></p></Label
              >
              <Input
                id="angle-z"
                v-model.number="sceneStates.cameras[selectedCamId]!.rotation.z"
                :disabled="
                  sceneStates.cameras[selectedCamId]!.isLockingRotation ||
                  props.workspace == null
                "
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
          </div>

          <div class="flex items-center gap-2">
            <input
              id="lock-rotation"
              v-model="sceneStates.cameras[selectedCamId]!.isLockingRotation"
              :disabled="props.workspace == null"
              type="checkbox"
              @change="onToggleLockRotation"
            />
            <label for="lock-rotation">Lock Rotation</label>
          </div>

          <div>
            <Label><p>Aspect Ratio</p></Label>
            <div class="flex flex-row gap-2 justify-center items-center">
              <Input
                id="aspect-ratio-width"
                v-model.number="sceneStates.cameras[selectedCamId]!.aspectWidth"
                :disabled="props.workspace == null"
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
              <p>:</p>
              <Input
                id="aspect-ratio-height"
                v-model.number="
                  sceneStates.cameras[selectedCamId]!.aspectHeight
                "
                :disabled="props.workspace == null"
                disabled-class="disabled-input"
                type="number"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
          </div>

          <div>
            <Label for="fov">Vertical Field of View</Label>
            <Input
              id="fov"
              v-model.number="sceneStates.cameras[selectedCamId]!.fov"
              :disabled="props.workspace == null"
              disabled-class="disabled-input"
              type="number"
              min="10"
              max="180"
              @change="sceneStates.markedForCheck.add(selectedCamId)"
            />
          </div>

          <div class="grid grid-flow-row grid-cols-2 gap-2">
            <Button
              size="sm"
              variant="ghost"
              :disabled="isLockingPosition"
              disabled-class="disabled-input"
              @click="
                sceneStates.cameras[selectedCamId]!.isHidingArrows =
                  !sceneStates.cameras[selectedCamId]!.isHidingArrows;
                sceneStates.markedForCheck.add(selectedCamId);
              "
            >
              <template
                v-if="sceneStates.cameras[selectedCamId]!.isLockingPosition"
              >
                <LockKeyhole class="h-3 w-3" />
              </template>

              <template v-else>
                <Eye
                  v-if="!sceneStates.cameras[selectedCamId]!.isHidingArrows"
                  class="h-3 w-3"
                />
                <EyeOff v-else class="h-3 w-3" />
              </template>
              Arrows
            </Button>

            <Button
              size="sm"
              variant="ghost"
              :disabled="
                sceneStates.cameras[selectedCamId]!.isLockingRotation ||
                props.workspace == null
              "
              disabled-class="disabled-input"
              @click="
                sceneStates.cameras[selectedCamId]!.isHidingWheels =
                  !sceneStates.cameras[selectedCamId]!.isHidingWheels;
                sceneStates.markedForCheck.add(selectedCamId);
              "
            >
              <template
                v-if="sceneStates.cameras[selectedCamId]!.isLockingRotation"
              >
                <LockKeyhole class="h-3 w-3" />
              </template>

              <template v-else>
                <Eye
                  v-if="!sceneStates.cameras[selectedCamId]!.isHidingWheels"
                  class="h-3 w-3"
                />
                <EyeOff v-else class="h-3 w-3" />
              </template>
              Wheels
            </Button>

            <Button
              size="sm"
              variant="ghost"
              :disabled="sceneStates.currentCamId.value == selectedCamId"
              @click="sceneStates.cameraManagement.switchToCam(selectedCamId)"
            >
              <Eye class="h-3 w-3" />
              Preview
            </Button>

            <Button
              size="sm"
              variant="ghost"
              :disabled="
                sceneStates.currentCamId.value == selectedCamId ||
                props.workspace == null
              "
              @click="
                deleteCamera(selectedCamId);
                sceneStates.markedForCheck.add(selectedCamId);
              "
            >
              <Trash2 class="h-3 w-3" />
              Delete
            </Button>

            <Button
              size="sm"
              variant="outline"
              class="flex-1"
              :disabled="
                sceneStates.cameras[selectedCamId]!.isLockingPosition ||
                sceneStates.cameras[selectedCamId]!.isLockingRotation ||
                props.workspace == null
              "
              disabled-class="disabled-input"
              @click="
                moveCameraHere(selectedCamId!);
                sceneStates.markedForCheck.add(selectedCamId);
              "
            >
              Move Here
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Frustum Properties -->
      <Card v-if="selectedCamId && sceneStates.cameras[selectedCamId]">
        <CardHeader
          class="cursor-pointer flex items-center justify-between"
          @click="isFrustumPropertiesOpen = !isFrustumPropertiesOpen"
          @change="sceneStates.markedForCheck.add(selectedCamId)"
        >
          <CardTitle class="text-base flex items-center gap-2">
            <Pyramid class="h-4 w-4" />
            Frustum Properties
          </CardTitle>
          <span class="text-sm">
            <ChevronDown
              v-if="isFrustumPropertiesOpen"
              class="inline h-4 w-4" />
            <ChevronLeft v-else class="inline h-4 w-4"
          /></span>
        </CardHeader>
        <CardContent v-if="isFrustumPropertiesOpen" class="space-y-2">
          <div class="grid grid-cols-2 gap-2">
            <Button
              size="sm"
              class="flex-1"
              @click="
                sceneStates.cameras[selectedCamId]!.isHidingFrustum =
                  !sceneStates.cameras[selectedCamId]!.isHidingFrustum;
                sceneStates.markedForCheck.add(selectedCamId);
              "
            >
              <Eye
                v-if="!sceneStates.cameras[selectedCamId]!.isHidingFrustum"
                class="h-3 w-3"
              />
              <EyeOff v-else class="h-3 w-3" />
              Frustum
            </Button>
            <Button
              size="sm"
              class="flex-1"
              variant="outline"
              @click="randomNewFrustumColor()"
            >
              <Dices class="h-3 w-3" />
              Random Color
            </Button>
          </div>
          <div class="grid grid-cols-3 gap-2">
            <div>
              <Label for="color-r"><p>R</p></Label>
              <Input
                id="color-r"
                v-model.number="
                  sceneStates.cameras[selectedCamId]!.frustumColor.r
                "
                type="number"
                min="0"
                max="1"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
            <div>
              <Label for="color-g"><p>G</p></Label>
              <Input
                id="color-g"
                v-model.number="
                  sceneStates.cameras[selectedCamId]!.frustumColor.g
                "
                type="number"
                min="0"
                max="1"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
            <div>
              <Label for="color-b"><p>B</p></Label>
              <Input
                id="color-b"
                v-model.number="
                  sceneStates.cameras[selectedCamId]!.frustumColor.b
                "
                type="number"
                min="0"
                max="1"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-2">
            <div>
              <Label for="opacity">Opacity</Label>
              <Input
                id="opacity"
                v-model.number="
                  sceneStates.cameras[selectedCamId]!.frustumColor.a
                "
                type="number"
                min="0"
                max="1"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
            <div>
              <Label for="length">Length</Label>
              <Input
                id="length"
                v-model.number="
                  sceneStates.cameras[selectedCamId]!.frustumLength
                "
                type="number"
                min="0"
                max="1e6"
                @change="sceneStates.markedForCheck.add(selectedCamId)"
              />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<style lang="css" scoped>
input {
  field-sizing: content;
}

/* For WebKit browsers (Chrome, Safari) */
input[type="number"]::-webkit-outer-spin-button,
input[type="number"]::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0; /* Important for removing extra space */
}

/* For Mozilla Firefox */
input[type="number"] {
  appearance: textfield;
  -moz-appearance: textfield;
}

input {
  border-radius: 5px;
  border: 1px solid black;
  outline: 1px solid white;
  box-sizing: border-box;

  margin-top: 4px;
}

.disabled-input {
  background-color: var(--color-gray-200);
  cursor: not-allowed;
}
</style>
