<script setup lang="ts">
import { ref } from "vue";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import Button from "./ui/button/Button.vue";
import Input from "./ui/input/Input.vue";
import Label from "./ui/label/Label.vue";
import Badge from "./ui/badge/Badge.vue";
import {
  Camera,
  Trash2,
  Eye,
  Settings,
  ChevronLeft,
  ChevronDown,
  MapPinPlus,
} from "lucide-vue-next";
import { SCENE_STATES_KEY } from "./3d/scene-states-provider/create-scene-states";
import type { ICamera } from "~/types/camera";

const sceneStates = inject(SCENE_STATES_KEY)!;

const currentCamData = ref<ICamera | null>(null);

watch(
  sceneStates.currentCam,
  (currentCam) => {
    currentCamData.value = sceneStates.cameras[currentCam!] ?? null;
  },
  {
    once: false,
  },
);

const isCameraPropertiesOpen = ref(true);

// const addCamera = () => {
//   const newCamera: CameraData = {
//     id: `cam${Date.now()}`,
//     name: `Camera ${cameras.value.length + 1}`,
//     position: { x: 5, y: 5, z: 5 },
//     rotation: { x: 0, y: 0, z: 0 },
//     fov: 60,
//   };
//   cameras.value.push(newCamera);
// };

const deleteCamera = (id: string) => {
  // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
  delete sceneStates.cameras[id];
};

// const updateCamera = (id: string, updates: Partial<CameraData>) => {
//   cameras.value = cameras.value.map((cam) =>
//     cam.id === id ? { ...cam, ...updates } : cam,
//   );
// };
</script>

<template>
  <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        <Camera class="h-5 w-5" />
        Camera Gallery
      </h2>
      <Button size="sm" @click="sceneStates.cameraManagement.spawnCameraHere">
        <MapPinPlus class="h-4 w-4" />
      </Button>
      <!-- <Button size="sm" @click="addCamera">
        <Plus class="h-4 w-4" />
      </Button> -->
    </div>

    <!-- Camera Dropdown -->
    <div class="mb-3">
      <Label for="camera-select" class="mb-1 block">Select Camera</Label>
      <select
        id="camera-select"
        v-model="sceneStates.currentCam"
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
      <div class="flex gap-2 mt-2">
        <Button size="sm" variant="ghost" :disabled="!sceneStates.currentCam">
          <Eye class="h-3 w-3" />
          Preview
        </Button>
        <Button
          size="sm"
          variant="ghost"
          :disabled="!sceneStates.currentCam"
          @click="deleteCamera(sceneStates.currentCam.value!)"
        >
          <Trash2 class="h-3 w-3" />
          Delete
        </Button>
      </div>
    </div>

    <div class="mb-3">
      <Badge variant="secondary" class="w-full justify-center">
        {{ Object.keys(sceneStates.cameras).length }} Camera{{
          Object.keys(sceneStates.cameras).length !== 1 ? "s" : ""
        }}
        Active
      </Badge>
    </div>

    <!-- Camera Properties -->
    <Card v-if="currentCamData">
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
      <CardContent v-if="isCameraPropertiesOpen" class="space-y-4">
        <div>
          <Label for="camera-name">Name</Label>
          <Input id="camera-name" v-model="currentCamData.name" />
        </div>

        <div class="grid grid-cols-3 gap-2">
          <div>
            <Label for="pos-x">X</Label>
            <Input
              id="pos-x"
              v-model.number="currentCamData.position.x"
              type="number"
            />
          </div>
          <div>
            <Label for="pos-y">Y</Label>
            <Input
              id="pos-y"
              v-model.number="currentCamData.position.y"
              type="number"
            />
          </div>
          <div>
            <Label for="pos-z">Z</Label>
            <Input
              id="pos-z"
              v-model.number="currentCamData.position.z"
              type="number"
            />
          </div>
        </div>

        <div>
          <Label for="fov">Field of View</Label>
          <Input
            id="fov"
            v-model.number="currentCamData.fov"
            type="number"
            min="10"
            max="120"
          />
        </div>

        <div class="flex gap-2">
          <Button size="sm" class="flex-1">
            <Eye class="h-3 w-3 mr-2" />
            Preview POV
          </Button>
          <Button size="sm" variant="outline" class="flex-1">
            Place in Scene
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
