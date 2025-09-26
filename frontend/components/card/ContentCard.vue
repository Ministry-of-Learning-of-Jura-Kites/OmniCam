<script setup lang="ts">
const props = defineProps<{
  name: string;
  description: string;
  imagePath: string;
  redirectLink: string;
}>();

const emit = defineEmits<{
  (e: "update" | "delete" | "updateImage", file?: File): void;
}>();

function handleUpdate() {
  emit("update");
}

function handleDelete() {
  emit("delete");
}

const showFull = ref<boolean>(false);

// Triggered when the upload icon is clicked
function handleUploadClick() {
  const fileInput = document.getElementById(
    `file-input-${props.name}`,
  ) as HTMLInputElement;
  if (fileInput) fileInput.click();
}

// Triggered when user selects a file
function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  if (target.files?.[0]) {
    emit("updateImage", target.files[0]); // send file to parent
  }
}
</script>

<template>
  <Card :class="['w-2xs', showFull ? 'max-h-none' : 'max-h-[400px]']">
    <CardHeader class="relative p-0 w-full">
      <router-link :to="props.redirectLink" class="w-full overflow-hidden">
        <template v-if="props.imagePath">
          <img
            :src="props.imagePath"
            alt="Project/Model Image"
            class="w-full min-h-[250px] max-h-[250px] object-cover rounded-t-lg"
          />
        </template>
        <template v-else>
          <div
            class="w-full min-h-[250px] max-h-[250px] flex items-center justify-center bg-gray-200 text-gray-700 text-6xl font-bold rounded-t-lg"
          >
            {{ props.name }}
          </div>
        </template>
      </router-link>

      <i
        class="fa fa-upload absolute top-2 right-2 p-2 rounded-full cursor-pointer"
        @click="handleUploadClick"
      ></i>
      <input
        :id="`file-input-${props.name}`"
        type="file"
        accept=".jpg,.png"
        class="hidden"
        @change="handleFileChange"
      />
    </CardHeader>

    <CardContent class="p-0 mt-[-20px]">
      <p class="pl-2 text-lg font-semibold break-words">{{ props.name }}</p>
      <p
        class="pl-2 text-sm text-gray-600 mt-1 break-words"
        :class="{ 'line-clamp-1': !showFull }"
      >
        {{ props.description }}
      </p>

      <button
        v-if="props.description.length > 100"
        class="text-blue-500 text-sm mt-1 underline"
        @click="showFull = !showFull"
      >
        {{ showFull ? "See Less" : "See More" }}
      </button>
    </CardContent>

    <CardFooter class="flex justify-end gap-2 p-4 mb-0.5">
      <i
        class="fa fa-pencil text-blue-500 cursor-pointer hover:text-blue-600"
        @click="handleUpdate"
      ></i>
      <i
        class="fa fa-trash text-red-500 cursor-pointer hover:text-red-600"
        @click="handleDelete"
      ></i>
    </CardFooter>
  </Card>
</template>
