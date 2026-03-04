import type { SceneStatesWithHelper } from "~/types/scene-states";

export const MODEL_INFO_KEY = "model_info";

export const IS_PANEL_OPEN_KEY: InjectionKey<boolean> = Symbol("is_panel_open");

export const TOGGLE_PANEL_KEY: InjectionKey<() => void> =
  Symbol("toggle_panel");

export const TOGGLE_MINIMAP_KEY: InjectionKey<() => void> =
  Symbol("toggle_map");

export const IS_MAP_OPEN_KEY: InjectionKey<boolean> = Symbol("is_map_open");

export const IS_CALIBRATING_KEY: InjectionKey<boolean> =
  Symbol("is_calibrating");

export const TOGGLE_CALIBRATION_KEY: InjectionKey<() => void> =
  Symbol("toggle_calibration");

export const SCENE_STATES_KEY: InjectionKey<SceneStatesWithHelper> =
  Symbol("3d-scene-states");

export const CALIBRATION_GRID_SCALE: InjectionKey<Ref<number>> = Symbol(
  "calibration_grid_scale",
);

export const CALIBRATION_SCALE: InjectionKey<Ref<number>> =
  Symbol("calibration_scale");

export const CALIBRATION_HEIGHT: InjectionKey<Ref<number>> =
  Symbol("calibration_height");
