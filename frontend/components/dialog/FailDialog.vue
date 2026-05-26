<script setup lang="ts">
const props = defineProps<{
  open: boolean;
  icon?: string;
  message: string;
  haveReturnButton?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "close-all"): void;
  // eslint-disable-next-line @typescript-eslint/unified-signatures
  (e: "return-to-form"): void;
  // eslint-disable-next-line @typescript-eslint/unified-signatures
  (e: "go-back"): void;
}>();

function close() {
  emit("close-all");
}

function returnToForm() {
  emit("return-to-form");
}

function goBack() {
  emit("go-back");
}
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent
      class="z-9999 max-w-md w-full h-[350px] rounded-2xl p-6 shadow-lg flex flex-col items-center justify-center space-y-1"
    >
      <i
        :class="props.icon || 'i-mdi-alert-circle'"
        class="text-red-400 text-8xl"
      ></i>
      <h2 class="text-xl font-semibold text-center text-red-300">
        {{ props.message }}
      </h2>

      <DialogFooter class="w-full flex justify-center! items-center">
        <DialogClose as-child>
          <Button
            v-if="props.haveReturnButton"
            type="button"
            variant="destructive"
            @click="returnToForm"
          >
            Back to Form
          </Button>

          <DialogClose as-child>
            <Button type="button" variant="secondary" @click="goBack">
              Go Back
            </Button>
          </DialogClose>
        </DialogClose>

        <DialogClose as-child>
          <Button type="button" variant="outline" @click="close">
            Close
          </Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
