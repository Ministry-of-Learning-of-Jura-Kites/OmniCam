<script setup lang="ts">
import { ref, watch, shallowRef } from "vue";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "~/lib/ui";
import { Check, ChevronsUpDown } from "lucide-vue-next";
import Button from "@/components/ui/button/Button.vue";
import Papa from "papaparse";

type Camerapreset = {
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

const props = defineProps<{
  modelValue: boolean;
  onConfirm: (preset: Camerapreset) => void;
}>();

const emit = defineEmits(["update:modelValue"]);

const presets = shallowRef<Camerapreset[]>([]);
const selectedPreset = ref<Camerapreset | null>(null);
const isLoading = ref(false);
const isLoaded = ref(false);
const openCombobox = ref(false);

watch(
  () => props.modelValue,
  async (open) => {
    if (open && !isLoaded.value) {
      await loadCsv();
    } else if (!open) {
      selectedPreset.value = null;
      openCombobox.value = false;
    }
  },
);

async function loadCsv() {
  try {
    isLoading.value = true;
    const resp = await fetch("/data/camera_parameter.csv");
    const text = await resp.text();
    const { data } = Papa.parse<Camerapreset>(text, { header: true });
    presets.value = data
      .filter((row) => row.vendor && row.camera)
      .map((row, index) => ({
        ...row,
        _id: `${index}-${row.vendor}-${row.camera}`,
      }));

    isLoaded.value = true;
  } finally {
    isLoading.value = false;
  }
}

const gcd = (a: number, b: number): number => {
  return b === 0 ? a : gcd(b, a % b);
};

const displayAspectRatio = computed(() => {
  const preset = selectedPreset.value;
  if (!preset) return "";

  const match = preset.sensor_name.match(/(\d+(?:\.\d+)?):(\d+(?:\.\d+)?)/);
  if (match) {
    return `(${match[0]})`;
  }

  const w = parseInt(preset.res_w);
  const h = parseInt(preset.res_h);

  if (!w || !h) return "";

  const divisor = gcd(w, h);
  return `(${w / divisor}:${h / divisor})`;
});

function handleSelect(preset: Camerapreset) {
  selectedPreset.value = preset;
  openCombobox.value = false;
}

function confirm() {
  if (!selectedPreset.value) return;
  props.onConfirm(selectedPreset.value);
  selectedPreset.value = null;
  emit("update:modelValue", false);
}

function close() {
  emit("update:modelValue", false);
}
</script>

<template>
  <Dialog :open="modelValue" @update:open="emit('update:modelValue', $event)">
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>Select Camera</DialogTitle>
      </DialogHeader>

      <div class="space-y-4">
        <div class="flex flex-col gap-2">
          <label
            class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Camera Model
          </label>

          <Popover v-model:open="openCombobox">
            <PopoverTrigger as-child>
              <Button
                variant="outline"
                role="combobox"
                :aria-expanded="openCombobox"
                class="w-full justify-between font-normal"
                :disabled="isLoading"
              >
                <span v-if="isLoading">Loading data...</span>
                <span v-else-if="selectedPreset" class="truncate">
                  {{ selectedPreset.vendor }} - {{ selectedPreset.camera }} ({{
                    selectedPreset.sensor_name
                  }})
                </span>
                <span v-else class="text-muted-foreground"
                  >Select camera...</span
                >
                <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
              </Button>
            </PopoverTrigger>

            <PopoverContent
              class="p-0"
              :style="{ width: 'var(--radix-popover-trigger-width)' }"
              align="start"
            >
              <Command
                :filter-function="
                  (list: any[], search: string) =>
                    list.filter((i) =>
                      i.toLowerCase().includes(search.toLowerCase()),
                    )
                "
              >
                <CommandInput placeholder="Search camera model..." />
                <CommandEmpty>No camera found.</CommandEmpty>
                <CommandList class="max-h-[300px] overflow-y-auto">
                  <CommandGroup>
                    <CommandItem
                      v-for="preset in presets"
                      :key="preset._id"
                      :value="`${preset.vendor} ${preset.camera} ${preset.sensor_name}`"
                      @select="handleSelect(preset)"
                    >
                      <Check
                        :class="
                          cn(
                            'mr-2 h-4 w-4',
                            selectedPreset?._id === preset._id
                              ? 'opacity-100'
                              : 'opacity-0',
                          )
                        "
                      />

                      <div class="flex flex-col">
                        <span> {{ preset.vendor }} - {{ preset.camera }} </span>
                        <span class="text-xs text-muted-foreground">
                          {{ preset.sensor_name }}
                        </span>
                      </div>
                    </CommandItem>
                  </CommandGroup>
                </CommandList>
              </Command>
            </PopoverContent>
          </Popover>
        </div>

        <div
          v-if="selectedPreset"
          class="border p-3 rounded-md bg-muted/30 text-sm space-y-1"
        >
          <div class="grid grid-cols-2 gap-x-4 gap-y-1">
            <p>
              <span class="font-semibold">Vendor:</span>
              {{ selectedPreset.vendor }}
            </p>
            <p>
              <span class="font-semibold">Camera:</span>
              {{ selectedPreset.camera }}
            </p>
            <p class="col-span-2">
              <span class="font-semibold">Sensor:</span>
              {{ selectedPreset.sensor_name }}
            </p>
          </div>

          <hr class="my-2" />

          <div class="grid grid-cols-2 gap-x-4 gap-y-1">
            <p>
              <span class="font-semibold">Aspect:</span>
              {{ selectedPreset.aspect }}
            </p>
            <p>
              <span class="font-semibold">V-FOV:</span>
              {{ selectedPreset.fov }}°
            </p>

            <p class="col-span-2">
              <span class="font-semibold">Resolution:</span>
              {{ selectedPreset.res_w }} × {{ selectedPreset.res_h }}
              <span>
                {{ displayAspectRatio }}
              </span>
            </p>

            <p class="col-span-2">
              <span class="font-semibold">Sensor Size:</span>
              {{ selectedPreset.sensor_w_mm }}mm ×
              {{ selectedPreset.sensor_h_mm }}mm
            </p>

            <p>
              <span class="font-semibold">Pitch:</span>
              {{ selectedPreset.pixel_pitch }} mm
            </p>
            <p>
              <span class="font-semibold">Focal:</span>
              {{ selectedPreset.focal_length }} mm
            </p>
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button variant="ghost" @click="close()">Cancel</Button>
        <Button :disabled="!selectedPreset" @click="confirm()">Confirm</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
:deep([cmdk-list]) {
  scrollbar-width: thin;
}
</style>
