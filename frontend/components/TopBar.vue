<script setup>
import Button from "./ui/button/Button.vue";
import Badge from "./ui/badge/Badge.vue";
import Card from "./ui/card/Card.vue";

import {
  Play,
  Square,
  RotateCcw,
  Maximize,
  Download,
  Upload,
  Save,
  User,
  LogOut,
} from "lucide-vue-next";
</script>

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
          <Badge variant="secondary" class="ml-2"> Active </Badge>
        </div>
      </Card>
    </div>

    <!-- Scene Controls -->
    <div class="flex items-center gap-2">
      <Button size="sm" variant="outline">
        <RotateCcw class="h-4 w-4 mr-2" />
        Reset View
      </Button>

      <Button size="sm" variant="outline">
        <Maximize class="h-4 w-4 mr-2" />
        Fullscreen
      </Button>

      <div class="h-6 w-px bg-border mx-2"></div>

      <Button size="sm">
        <Upload class="h-4 w-4 mr-2" />
        Upload Model
      </Button>

      <Button size="sm" variant="outline">
        <Save class="h-4 w-4 mr-2" />
        Save
      </Button>

      <Button size="sm" variant="outline">
        <Download class="h-4 w-4 mr-2" />
        Export
      </Button>
    </div>

    <!-- User Actions -->
    <div class="flex items-center gap-2">
      <Button size="sm" variant="ghost">
        <User class="h-4 w-4 mr-2" />
        Profile
      </Button>
      <Button size="sm" variant="ghost">
        <LogOut class="h-4 w-4" />
      </Button>
    </div>
  </div>
</template>
