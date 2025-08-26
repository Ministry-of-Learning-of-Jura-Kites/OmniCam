<script setup lang="ts">
import { ref, reactive, computed } from "vue";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import Button from "./ui/button/Button.vue";
import Input from "./ui/input/Input.vue";
import Label from "./ui/label/Label.vue";
import Badge from "./ui/badge/Badge.vue";
import { Camera, Plus, Trash2, Eye, Settings } from "lucide-vue-next";

interface CameraData {
  id: string;
  name: string;
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
  fov: number;
}

const cameras = ref<CameraData[]>([
  {
    id: "placeholderCam",
    name: "Placeholder View",
    position: { x: 10, y: 10, z: 10 },
    rotation: { x: 0, y: 0, z: 0 },
    fov: 60,
  },
]);

const selectedCamera = ref<string | null>("placeholderCam");

const addCamera = () => {
  const newCamera: CameraData = {
    id: `cam${Date.now()}`,
    name: `Camera ${cameras.value.length + 1}`,
    position: { x: 5, y: 5, z: 5 },
    rotation: { x: 0, y: 0, z: 0 },
    fov: 60,
  };
  cameras.value.push(newCamera);
};

const deleteCamera = (id: string) => {
  cameras.value = cameras.value.filter((cam) => cam.id !== id);
  if (selectedCamera.value === id) {
    selectedCamera.value = cameras.value[0]?.id || null;
  }
};

const updateCamera = (id: string, updates: Partial<CameraData>) => {
  cameras.value = cameras.value.map((cam) =>
    cam.id === id ? { ...cam, ...updates } : cam,
  );
};

const selectedCameraData = computed(() =>
  cameras.value.find((cam) => cam.id === selectedCamera.value),
);
</script>

<template>
  <div class="w-80 bg-card border-l border-border p-4 overflow-y-auto">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        <Camera class="h-5 w-5" />
        Camera Gallery
      </h2>
      <Button size="sm" @click="addCamera">
        <Plus class="h-4 w-4" />
      </Button>
    </div>

    <!-- Camera List -->
    <div class="space-y-2 mb-6">
      <Card
        v-for="camera in cameras"
        :key="camera.id"
        :class="[
          'cursor-pointer transition-colors',
          selectedCamera === camera.id
            ? 'bg-survey-surface border-primary'
            : 'hover:bg-survey-surface-hover',
        ]"
        @click="selectedCamera = camera.id"
      >
        <CardContent class="p-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="font-medium">{{ camera.name }}</div>
              <div class="text-sm text-muted-foreground">
                FOV: {{ camera.fov }}°
              </div>
            </div>
            <div class="flex gap-1">
              <Button size="sm" variant="ghost">
                <Eye class="h-3 w-3" />
              </Button>
              <Button
                size="sm"
                variant="ghost"
                @click.stop="deleteCamera(camera.id)"
              >
                <Trash2 class="h-3 w-3" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Camera Properties -->
    <Card v-if="selectedCameraData">
      <CardHeader>
        <CardTitle class="text-base flex items-center gap-2">
          <Settings class="h-4 w-4" />
          Camera Properties
        </CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <div>
          <Label for="camera-name">Name</Label>
          <Input id="camera-name" v-model="selectedCameraData.name" />
        </div>

        <div class="grid grid-cols-3 gap-2">
          <div>
            <Label for="pos-x">X</Label>
            <Input
              id="pos-x"
              type="number"
              v-model.number="selectedCameraData.position.x"
            />
          </div>
          <div>
            <Label for="pos-y">Y</Label>
            <Input
              id="pos-y"
              type="number"
              v-model.number="selectedCameraData.position.y"
            />
          </div>
          <div>
            <Label for="pos-z">Z</Label>
            <Input
              id="pos-z"
              type="number"
              v-model.number="selectedCameraData.position.z"
            />
          </div>
        </div>

        <div>
          <Label for="fov">Field of View</Label>
          <Input
            id="fov"
            type="number"
            min="10"
            max="120"
            v-model.number="selectedCameraData.fov"
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

    <div class="mt-6">
      <Badge variant="secondary" class="w-full justify-center">
        {{ cameras.length }} Camera{{ cameras.length !== 1 ? "s" : "" }} Active
      </Badge>
    </div>
  </div>
</template>
