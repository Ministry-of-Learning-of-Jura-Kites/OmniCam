import {
  type Camera,
  type AutosaveEvent,
  AutosaveMessage,
  AutosaveResponse,
} from "~/messages/protobufs/autosave_event";
import type { ICamera } from "~/types/camera";
import type { SceneStates } from "~/types/scene-states";
import { Quaternion, ShaderMaterial } from "three";
import Color4 from "three/src/renderers/common/Color4.js";

function isEqual(a: Camera, b: Camera): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export function transformCameraToProtoEvent(cam: ICamera): Omit<Camera, "id"> {
  const quaternion = new Quaternion();
  quaternion.setFromEuler(cam.rotation);
  return {
    name: cam.name,
    angleX: quaternion.x,
    angleY: quaternion.y,
    angleZ: quaternion.z,
    angleW: quaternion.w,
    posX: cam.position.x,
    posY: cam.position.y,
    posZ: cam.position.z,
    widthRes: cam.widthRes,
    heightRes: cam.heightRes,
    fov: cam.fov,
    frustumColor: cam.frustumColor,
    frustumLength: cam.frustumLength,
    isHidingArrows: cam.isHidingArrows,
    isHidingWheels: cam.isHidingWheels,
    isLockingPosition: cam.isLockingPosition,
    isLockingRotation: cam.isLockingRotation,
    isHidingFrustum: cam.isHidingFrustum,
    distortion: structuredClone(toRaw(cam.distortion)),
  };
}

export function transformCameraToProtoEventWithId(
  camId: string,
  cam: ICamera,
): Camera {
  return { ...transformCameraToProtoEvent(cam), id: camId };
}

const depthMaterial = new ShaderMaterial({
  uniforms: {
    uNear: { value: 0.1 },
    uFar: { value: 50 },
  },
  vertexShader: `
        varying float vViewZ;
        void main() {
            vec4 mvPosition = modelViewMatrix * vec4(position, 1.0);
            vViewZ = -mvPosition.z; // Distance from camera
            gl_Position = projectionMatrix * mvPosition;
        }
    `,
  fragmentShader: `
        varying float vViewZ;
        uniform float uNear;
        uniform float uFar;
        void main() {
            // Linear interpolation
            float depth = (vViewZ - uNear) / (uFar - uNear);
            depth = clamp(depth, 0.0, 1.0);
            
            // Invert so Near is White (1.0) and Far is Black (0.0)
            float finalDepth = depth;
            
            gl_FragColor = vec4(vec3(finalDepth), 1.0);
        }
    `,
});

export function useAutosave(
  sceneStates: SceneStates,
  workspace: string | null,
) {
  if (workspace == null) return;

  const lastSynced: Map<string, Camera> = new Map(
    Object.entries(sceneStates.cameras!).map(([camId, cam]) => [
      camId,
      transformCameraToProtoEventWithId(camId, cam),
    ]),
  );

  const isServerUpdate = ref(false);

  onMounted(() => {
    watch(
      () => sceneStates.websocket?.data.value,
      async (messageBlob) => {
        if (!messageBlob) return;
        const buf = await (messageBlob as Blob).arrayBuffer();
        const resp = AutosaveResponse.decode(new Uint8Array(buf));

        sceneStates.lastSyncedVersion.value = resp.lastUpdatedVersion;

        // if (resp.calibrationAck) {
        //   isServerUpdate.value = true;
        //   sceneStates.calibrationScale.value = resp.calibrationAck.scaleFactor;
        //   sceneStates.calibrationHeight.value = resp.calibrationAck.modelHeight;
        //   sceneStates.calibrationVersion.value =
        //     resp.calibrationAck.lastUpdatedVersion;
        //   await nextTick();
        //   isServerUpdate.value = false;
        // }
      },
    );

    watch(
      () => [
        sceneStates.calibration.scale,
        sceneStates.calibration.heightOffset,
      ],
      ([newScale, newHeight], [oldScale, oldHeight]) => {
        if (isServerUpdate.value) return;
        if (newScale !== oldScale || newHeight !== oldHeight) {
          sceneStates.calibration.dirty = true;
        }
      },
    );

    watch(sceneStates.tresContext, (ctx) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (window as any)["captureDepthMap"] = () => {
        const renderer = ctx!.renderer.instance!;
        // 1. Store the original background/clear color
        const originalClearColor = new Color4();
        renderer.getClearColor(originalClearColor);
        const originalClearAlpha = renderer.getClearAlpha();

        const scene = ctx!.scene!;

        // 3. Temporarily override scene material
        const originalOverrideMaterial = scene.overrideMaterial;
        scene.overrideMaterial = depthMaterial;

        // 4. Set renderer to white background (to represent max distance)
        renderer.setClearColor(0x000000, 1);

        const camera = ctx!.camera.activeCamera;

        // 5. Render to a temporary canvas or the existing one
        renderer.render(scene, camera);

        // 6. Extract the data URL
        const dataURL = renderer.domElement.toDataURL("image/png");

        // 7. Restore scene state
        scene.overrideMaterial = originalOverrideMaterial;
        renderer.setClearColor(originalClearColor, originalClearAlpha);

        // Re-render the original scene so the canvas isn't stuck on the depth map
        renderer.render(scene, camera);

        // 8. Return or trigger download
        console.log("Depth map captured!");
        return dataURL;
      };
    });

    setInterval(() => {
      if (!sceneStates.websocket) return;

      const changed: AutosaveEvent[] = [];

      // Cameras
      if (sceneStates.markedForCheck.size > 0) {
        for (const camId of sceneStates.markedForCheck) {
          const prev = lastSynced.get(camId);
          const cam = sceneStates.cameras[camId];

          if (cam == undefined && prev == undefined) continue;

          if (cam == undefined) {
            lastSynced.delete(camId);
            changed.push({ delete: { id: camId } });
            continue;
          }

          const formattedCam = transformCameraToProtoEventWithId(camId, cam);
          if (prev == undefined || !isEqual(prev, formattedCam)) {
            changed.push({ upsert: { camera: formattedCam } });
            lastSynced.set(camId, formattedCam);
          }
        }
      }

      // Calibration
      if (sceneStates.calibration.dirty) {
        changed.push({
          calibrate: {
            scaleFactor: sceneStates.calibration.scale,
            modelHeight: sceneStates.calibration.heightOffset,
          },
        });
        sceneStates.calibration.dirty = false;
      }

      if (changed.length > 0) {
        sceneStates.localVersion.value += 1;
        const encoded = AutosaveMessage.encode({
          version: sceneStates.localVersion.value,
          events: changed,
        }).finish();
        sceneStates.websocket.send(encoded.buffer);
      }

      sceneStates.markedForCheck.clear();
    }, 2000);
  });
}
