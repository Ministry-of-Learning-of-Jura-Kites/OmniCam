import type { TresContext } from "@tresjs/core";
import type { Reactive } from "vue";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import type {
  SceneStates as BaseSceneStates,
  SceneStatesWithHelper,
} from "~/types/scene-states";
import * as THREE from "three";
import type { ICamera } from "~/types/camera";
import { useCameraManagement } from "../scene-3d/use-camera-management";
import { useSpectatorRotation } from "../scene-3d/use-spectator-rotation";
import { useSpectatorPosition } from "../scene-3d/use-spectator-position";
import { useWebSocket } from "@vueuse/core";
import type { RuntimeConfig } from "nuxt/schema";

interface ModelWithCamsResp {
  data: {
    modelId: string;
    cameras: Record<
      string,
      {
        name: string;
        angleX: number;
        angleY: number;
        angleZ: number;
        angleW: number;
        posX: number;
        posY: number;
        posZ: number;
        fov: number;
      }
    >;
  };
}

export const SCENE_STATES_KEY: InjectionKey<SceneStatesWithHelper> =
  Symbol("3d-scene-states");

async function loadCamsData(
  projectId: string,
  modelId: string,
  runtimeConfig: RuntimeConfig,
  workspace?: string,
): Promise<[Record<string, ICamera>, null] | [null, unknown]> {
  const workspaceSuffix = workspace == null ? "" : `/workspaces/${workspace}`;

  const fields = ["cameras"];

  const params = new URLSearchParams();
  for (const field of fields) {
    params.append("fields", field);
  }
  params.append("t", String(Date.now()));

  try {
    const rawResp = await fetch(
      `http://${runtimeConfig.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models/${modelId}${workspaceSuffix}?${params.toString()}`,
    );
    const resp: ModelWithCamsResp = await rawResp.json();
    if (rawResp.status == 404) {
      return [null, { action: "not-found" }];
    }

    return [
      Object.fromEntries(
        Object.entries(resp.data.cameras).map(([camId, rawCam]) => {
          const cam: ICamera = {
            name: rawCam.name,
            position: new THREE.Vector3(rawCam.posX, rawCam.posY, rawCam.posZ),
            rotation: new THREE.Euler().setFromQuaternion(
              new THREE.Quaternion(
                rawCam.angleX,
                rawCam.angleY,
                rawCam.angleZ,
                rawCam.angleW,
              ),
              "YXZ",
            ),
            fov: rawCam.fov,
            isHidingArrows: false,
            isHidingWheels: false,
          };
          return [camId, cam];
        }),
      ),
      null,
    ];
  } catch (e) {
    return [null, e];
  }
}

export async function createBaseSceneStates(
  projectId: string,
  modelId: string,
  runtimeConfig: RuntimeConfig,
  workspace?: string,
) {
  const tresContext = ref<TresContext | null>(null);

  const draggableObjects: Set<Obj3DWithUserData> = new Set();

  const isDraggingObject: Ref<boolean> = ref(false);

  const currentCamId: Ref<string | null> = ref(null);

  const spectatorCameraPosition: Reactive<THREE.Vector3> = reactive(
    new THREE.Vector3(0, 1, 0),
  );

  const spectatorCameraRotation: Reactive<THREE.Euler> = reactive(
    new THREE.Euler(0, 0, 0, "YXZ"),
  );

  const spectatorCameraFov: Ref<number> = ref(75);

  const currentCameraPosition: Ref<THREE.Vector3> = ref(
    spectatorCameraPosition,
  );

  const currentCameraRotation: Ref<THREE.Euler> = ref(spectatorCameraRotation);

  const tresCanvasParent: Ref<HTMLDivElement | null> = ref(null);

  const [camsData, error] = await loadCamsData(
    projectId,
    modelId,
    runtimeConfig,
    workspace,
  );

  if (error != null) {
    return { error: error as unknown };
  }

  const cameras = reactive<Record<string, ICamera>>(camsData!);

  const currentCam = computed(() => {
    return currentCamId.value == null ? null : cameras![currentCamId.value];
  });

  const currentCameraFov = computed({
    get() {
      if (currentCamId.value == null) {
        return spectatorCameraFov.value;
      }
      return currentCam.value?.fov;
    },
    set(newFOV) {
      if (currentCam.value) {
        currentCam.value.fov = newFOV!;
      }
    },
  });

  const websocketUrl = `ws://${runtimeConfig.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/projects/${projectId}/models/${modelId}/autosave`;

  const websocket = useWebSocket(websocketUrl, {
    autoReconnect: {
      delay: 1000,
      onFailed: () => {
        alert("Failed to connect websocket after multiple retries.");
      },
    },
  });

  const markedForCheck = reactive(new Set<string>());

  const sceneStates = {
    tresContext,
    draggableObjects,
    isDraggingObject,
    currentCamId,
    currentCam,
    currentCameraPosition,
    currentCameraRotation,
    currentCameraFov,
    spectatorCameraPosition,
    spectatorCameraRotation,
    spectatorCameraFov,
    tresCanvasParent,
    websocket,
    cameras,
    error: null,
    markedForCheck,
  } as const;

  // websocket.ws.value!.onclose = (_closeEvent: CloseEvent) => {
  //   sceneStates.websocket = useWebSocket(websocketUrl);
  // };

  return sceneStates;
}

export function createSceneStatesWithHelper(
  sceneStates: Awaited<BaseSceneStates>,
) {
  const sceneStatesWithCam = {
    ...sceneStates,
    cameraManagement: useCameraManagement(sceneStates),
    spectatorPosition: useSpectatorPosition(sceneStates),
    spectatorRotation: useSpectatorRotation(sceneStates),
  };
  return sceneStatesWithCam;
}
