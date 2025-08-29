import type {
  createBaseSceneStates,
  createSceneStatesWithHelper,
} from "~/components/3d/scene-states-provider/create-scene-states";

export type SceneStates = ReturnType<typeof createBaseSceneStates>;

export type SceneStatesWithHelper = ReturnType<
  typeof createSceneStatesWithHelper
>;
