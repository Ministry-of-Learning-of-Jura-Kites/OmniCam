<script setup lang="ts">
import type { ICamera } from "~/types/camera";

const props = defineProps<{
  camId: string;
  cam: ICamera;
  onUpdate: (id: string) => void;
}>();

watch(
  () => [
    props.cam.position.x,
    props.cam.position.y,
    props.cam.position.z,
    props.cam.rotation.x,
    props.cam.rotation.y,
    props.cam.rotation.z,
  ],
  () => {
    props.onUpdate(props.camId);
  },
  { flush: "sync" },
);
</script>
