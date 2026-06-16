import type { InjectionKey, Ref } from "vue";
import type { SceneStatesWithHelper } from "~/types/scene-states";

export const MODEL_INFO_KEY = "model_info";

// export const DEFAULT_TOOL_MODE: ToolMode = "orbit";

export interface MiniMapInfo {
  toggleMap: () => void;
  isMapOpen: Ref<boolean>;
}

export const MAP_KEY: InjectionKey<MiniMapInfo> = Symbol("map");

export const CAMERA_INSTANCE_KEY = Symbol("cameraInstance") as InjectionKey<{
  updateCameraMatrix: (id: string) => void;
}>;

export interface CamPanelInfo {
  selectedCamId: Ref<string | null>;
}

export interface CalibrationPanelInfo {
  toggleCalibration: () => void;
  calibrationGridScale: Ref<number>;
}

export interface SimulationPanelInfo {
  toggleSimulation: () => void;
}
export type PanelType =
  | "camera"
  | "calibration"
  | "measurement"
  | "algo"
  | "simulation";

// export type ToolMode = "none" | "orbit" | "measurement" | "calibration";

export interface PanelInfo {
  camPanelInfo: CamPanelInfo;

  calibrationPanelInfo: CalibrationPanelInfo;

  simulationPanelInfo: SimulationPanelInfo;

  currentPanel: Ref<PanelType>;

  // currentToolMode: Ref<ToolMode>;

  toggleAlgoPanel: () => void;

  toggleCameraPanel: () => void;

  toggleMeasurement: () => void;

  togglePanel: () => void;

  isPanelOpen: Ref<boolean>;
}

export const PANEL_KEY: InjectionKey<PanelInfo> = Symbol("panel");

export const SCENE_STATES_KEY: InjectionKey<
  Ref<SceneStatesWithHelper | undefined>
> = Symbol("3d-scene-states");

export const SCENE_STATES_READY_KEY: InjectionKey<Ref<boolean>> = Symbol(
  "3d-scene-states-ready",
);
