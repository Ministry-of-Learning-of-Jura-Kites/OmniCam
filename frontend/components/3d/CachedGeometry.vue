<script setup lang="ts">
const props = defineProps<{
  module: ModuleName;
  name: UniversalGeoName;
}>();

const { get } = useGeoCache();
const geometry = get(props.module, props.name);
</script>

<template>
  <primitive :object="geometry" />
</template>
