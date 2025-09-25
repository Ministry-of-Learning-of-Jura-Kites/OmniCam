<script setup lang="ts">
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";

const props = defineProps<{
  page: number;
  pageSize: number;
  totalItem: number;
}>();

// Define the custom event
const emit = defineEmits(["update:page"]);

// Function to handle page changes
const handlePageChange = (newPage: number) => {
  // Emit the new page value to the parent
  emit("update:page", newPage);
};
</script>

<template>
  <Pagination
    v-slot="{ page: currentPage }"
    :items-per-page="props.pageSize"
    :total="props.totalItem"
    :default-page="props.page"
  >
    <PaginationContent v-slot="{ items }">
      <PaginationPrevious @click="handlePageChange(currentPage - 1)" />

      <template v-for="(item, index) in items" :key="index">
        <PaginationItem
          v-if="item.type === 'page'"
          :value="item.value"
          :is-active="item.value === currentPage"
          @click="handlePageChange(item.value)"
        >
          {{ item.value }}
        </PaginationItem>

        <PaginationEllipsis
          v-else-if="item.type === 'ellipsis'"
          :index="index"
        />
      </template>

      <PaginationNext @click="handlePageChange(currentPage + 1)" />
    </PaginationContent>
  </Pagination>
</template>
