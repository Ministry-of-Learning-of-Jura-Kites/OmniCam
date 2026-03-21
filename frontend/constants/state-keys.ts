import type { SceneStatesWithHelper } from "~/types/scene-states";

export const MODEL_INFO_KEY = "model_info";

export const CURRENT_PANEL: InjectionKey<Ref<"camera" | "algo" | null>> =
  Symbol("current_panel");

export const TOGGLE_ALGO_PANEL_KEY: InjectionKey<() => void> =
  Symbol("toggle_algo_panel");

export const TOGGLE_CAM_PANEL_KEY: InjectionKey<() => void> =
  Symbol("toggle_panel");

export const TOGGLE_MINIMAP_KEY: InjectionKey<() => void> =
  Symbol("toggle_map");

export const IS_MAP_OPEN_KEY: InjectionKey<Ref<boolean>> =
  Symbol("is_map_open");

export const SCENE_STATES_KEY: InjectionKey<SceneStatesWithHelper> =
  Symbol("3d-scene-states");

export const CALIBRATION_GRID_SCALE: InjectionKey<Ref<number>> = Symbol(
  "calibration_grid_scale",
);
