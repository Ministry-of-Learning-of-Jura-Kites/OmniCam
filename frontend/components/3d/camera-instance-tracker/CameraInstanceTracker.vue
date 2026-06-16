<script setup lang="ts">
import { CAMERA_INSTANCE_KEY } from "~/constants/state-keys";
import type { ICamera } from "~/types/camera";

const props = defineProps<{
  camId: string;
  cam: ICamera;
  onUpdate: (id: string) => void;
}>();

const cameraInstance = inject(CAMERA_INSTANCE_KEY);

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
    cameraInstance?.updateCameraMatrix(props.camId);
  },
  { flush: "sync" },
);
</script>
