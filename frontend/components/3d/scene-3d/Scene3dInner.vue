<script setup lang="ts">
import { useTresContext } from "@tresjs/core";
import { tresContext } from "./refs";

onMounted(() => {
  const context = useTresContext();
  console.log("Tres context ready:", context.scene.value);
  tresContext.value = context;
});
</script>
