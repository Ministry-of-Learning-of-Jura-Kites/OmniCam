<!-- CameraInstanceTracker.vue -->
<script setup lang="ts">
import { CAMERA_INSTANCE_KEY } from "~/constants/state-keys";
import type { ICamera } from "~/types/camera";

const props = defineProps<{
  camId: string;
  cam: ICamera;
}>();

const cameraInstance = inject(CAMERA_INSTANCE_KEY)!;

watch(
  () => ({
    px: props.cam.position.x,
    py: props.cam.position.y,
    pz: props.cam.position.z,
    rx: props.cam.rotation.x,
    ry: props.cam.rotation.y,
    rz: props.cam.rotation.z,
  }),
  () => cameraInstance.updateCameraMatrix(props.camId),
);
</script>
