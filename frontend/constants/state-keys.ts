import type { InjectionKey, Ref } from "vue";
import type { SceneStatesWithHelper } from "~/types/scene-states";

export const MODEL_INFO_KEY = "model_info";

export interface MiniMapInfo {
  toggleMap: () => void;
  isMapOpen: Ref<boolean>;
}

export const MAP_KEY: InjectionKey<MiniMapInfo> = Symbol("map");

export interface CamPanelInfo {
  selectedCamId: Ref<string | null>;
}

export interface CalibrationPanelInfo {
  isCalibrating: Ref<boolean>;
  toggleCalibration: () => void;
  calibrationGridScale: Ref<number>;
}

export interface PanelInfo {
  camPanelInfo: CamPanelInfo;
  calibrationPanelInfo: CalibrationPanelInfo;
  currentPanel: Ref<"camera" | "algo">;
  toggleAlgoPanel: () => void;
  togglePanel: () => void;
  isPanelOpen: Ref<boolean>;
}
export const PANEL_KEY: InjectionKey<PanelInfo> = Symbol("panel");

export const SCENE_STATES_KEY: InjectionKey<SceneStatesWithHelper> =
  Symbol("3d-scene-states");
