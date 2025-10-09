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
} from "lucide-vue-next";
import { SCENE_STATES_KEY } from "./3d/scene-states-provider/create-scene-states";

const props = defineProps({
  workspace: {
    type: String,
    default: null,
  },
});

const sceneStates = inject(SCENE_STATES_KEY)!;

const selectedCamId = ref<string | null>(null);

const isCameraPropertiesOpen = ref(true);

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
  // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
  delete sceneStates.cameras[id];
};
</script>

<template>
  <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto">
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
        :disabled="props.workspace != null"
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
          {{ camera.name }} (FOV: {{ camera.fov }}°)
        </option>
      </select>
      <div class="flex gap-2 mt-2"></div>
    </div>

    <!-- <div class="mb-3">
      <Badge variant="secondary" class="w-full justify-center">
        {{ Object.keys(sceneStates.cameras).length }} Camera{{
          Object.keys(sceneStates.cameras).length !== 1 ? "s" : ""
        }}
        Active
      </Badge>
    </div> -->

    <!-- Camera Properties -->
    <Card v-if="selectedCamId && sceneStates.cameras[selectedCamId]">
      <CardHeader
        class="cursor-pointer flex items-center justify-between"
        @click="isCameraPropertiesOpen = !isCameraPropertiesOpen"
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
            @change="sceneStates.markedForCheck.add(selectedCamId)"
          />
        </div>

        <div class="grid grid-cols-3 gap-2">
          <div>
            <Label for="pos-x">X</Label>
            <Input
              id="pos-x"
              v-model.number="sceneStates.cameras[selectedCamId]!.position.x"
              :disabled="props.workspace == null"
              type="number"
              @change="sceneStates.markedForCheck.add(selectedCamId)"
            />
          </div>
          <div>
            <Label for="pos-y">Y</Label>
            <Input
              id="pos-y"
              v-model.number="sceneStates.cameras[selectedCamId]!.position.y"
              :disabled="props.workspace == null"
              type="number"
              @change="sceneStates.markedForCheck.add(selectedCamId)"
            />
          </div>
          <div>
            <Label for="pos-z">Z</Label>
            <Input
              id="pos-z"
              v-model.number="sceneStates.cameras[selectedCamId]!.position.z"
              :disabled="props.workspace == null"
              type="number"
              @change="sceneStates.markedForCheck.add(selectedCamId)"
            />
          </div>
        </div>

        <div class="grid grid-cols-3 gap-2">
          <div>
            <Label for="angle-x"
              ><p>θ<sub>x</sub></p></Label
            >
            <Input
              id="angle-x"
              v-model.number="sceneStates.cameras[selectedCamId]!.rotation.x"
              :disabled="props.workspace == null"
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
              :disabled="props.workspace == null"
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
              :disabled="props.workspace == null"
              type="number"
              @change="sceneStates.markedForCheck.add(selectedCamId)"
            />
          </div>
        </div>

        <div>
          <Label for="fov">Field of View</Label>
          <Input
            id="fov"
            v-model.number="sceneStates.cameras[selectedCamId]!.fov"
            :disabled="props.workspace == null"
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
            @click="
              sceneStates.cameras[selectedCamId]!.isHidingArrows =
                !sceneStates.cameras[selectedCamId]!.isHidingArrows;
              sceneStates.markedForCheck.add(selectedCamId);
            "
          >
            <Eye
              v-if="!sceneStates.cameras[selectedCamId]!.isHidingArrows"
              class="h-3 w-3"
            />
            <EyeOff v-else class="h-3 w-3" />
            Arrows
          </Button>

          <Button
            size="sm"
            variant="ghost"
            @click="
              sceneStates.cameras[selectedCamId]!.isHidingWheels =
                !sceneStates.cameras[selectedCamId]!.isHidingWheels;
              sceneStates.markedForCheck.add(selectedCamId);
            "
          >
            <Eye
              v-if="!sceneStates.cameras[selectedCamId]!.isHidingWheels"
              class="h-3 w-3"
            />
            <EyeOff v-else class="h-3 w-3" />
            Wheels
          </Button>

          <Button
            size="sm"
            variant="ghost"
            @click="sceneStates.cameraManagement.switchToCam(selectedCamId)"
          >
            <Eye class="h-3 w-3" />
            Preview
          </Button>

          <Button
            size="sm"
            variant="ghost"
            :disabled="!sceneStates.currentCamId || props.workspace == null"
            @click="
              deleteCamera(selectedCamId);
              sceneStates.markedForCheck.add(selectedCamId);
            "
          >
            <Trash2 class="h-3 w-3" />
            Delete
          </Button>

          <Button size="sm" class="flex-1">
            <Eye class="h-3 w-3 mr-2" />
            Frustum
          </Button>

          <Button
            size="sm"
            variant="outline"
            class="flex-1"
            @click="moveCameraHere(selectedCamId!)"
          >
            Move Here
          </Button>
        </div>
      </CardContent>
    </Card>
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
  width: 100%;
  border-radius: 5px;
  border: 1px solid black;
  outline: 1px solid white;
  box-sizing: border-box;
}
</style>
