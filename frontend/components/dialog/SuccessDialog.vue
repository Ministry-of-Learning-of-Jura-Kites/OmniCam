<script setup lang="ts">
const props = defineProps<{
  open: boolean;
  icon: string;
  message: string;
}>();

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
}>();
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent
      class="max-w-md w-full h-[350px] rounded-2xl p-6 shadow-lg flex flex-col items-center justify-center space-y-6"
    >
      <i :class="props.icon" class="text-green-500 text-8xl"></i>
      <h2 class="text-xl font-semibold text-center">
        {{ props.message }}
      </h2>

      <DialogFooter class="w-full flex !justify-center items-center">
        <DialogClose as-child>
          <Button type="button" variant="secondary"> Close </Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
