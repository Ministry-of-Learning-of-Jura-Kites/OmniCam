<script setup lang="ts">
import { useTresContext } from "@tresjs/core";
import { SCENE_STATES_KEY } from "../use-scene-state";

const sceneStates = inject(SCENE_STATES_KEY);
if (!sceneStates) throw new Error("SceneStates is missing!");

onMounted(() => {
  const context = useTresContext();
  console.log("Tres context ready:", context.scene.value);
  sceneStates.tresContext.value = context;
});
</script>
