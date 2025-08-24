<template>
  <div
    class="h-16 bg-card border-b border-border px-6 flex items-center justify-between"
  >
    <!-- Project Info -->
    <div class="flex items-center gap-4">
      <Card class="px-3 py-1 bg-survey-surface">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 bg-survey-accent rounded-full"></div>
          <span class="text-sm font-medium">Survey Project 01</span>
          <Badge variant="secondary" class="ml-2">Active</Badge>
        </div>
      </Card>
    </div>

    <!-- Scene Controls -->
    <div class="flex items-center gap-2">
      <Button variant="outline" size="sm">
        <RotateCcwIcon class="w-4 h-4 mr-2" />
        Reset View
      </Button>

      <Button variant="outline" size="sm">
        <MaximizeIcon class="w-4 h-4 mr-2" />
        Fullscreen
      </Button>

      <div class="h-6 w-px bg-border mx-2"></div>

      <Button size="sm">
        <UploadIcon class="w-4 h-4 mr-2" />
        Upload Model
      </Button>

      <Button variant="outline" size="sm">
        <SaveIcon class="w-4 h-4 mr-2" />
        Save
      </Button>

      <Button variant="outline" size="sm">
        <DownloadIcon class="w-4 h-4 mr-2" />
        Export
      </Button>
    </div>

    <!-- User Actions -->
    <div class="flex items-center gap-2">
      <Button variant="ghost" size="sm">
        <UserIcon class="w-4 h-4 mr-2" />
        Profile
      </Button>
      <Button variant="ghost" size="sm">
        <LogOutIcon class="w-4 h-4" />
      </Button>
    </div>
  </div>
</template>

<script setup>
import { Button, Card, Badge } from "@/components/ui";
import {
  RotateCcwIcon,
  MaximizeIcon,
  DownloadIcon,
  UploadIcon,
  SaveIcon,
  UserIcon,
  LogOutIcon,
} from "lucide-vue-next";
</script>
