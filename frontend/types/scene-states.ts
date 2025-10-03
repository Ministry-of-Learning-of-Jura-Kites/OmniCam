import type {
  createBaseSceneStates,
  createSceneStatesWithHelper,
} from "~/components/3d/scene-states-provider/create-scene-states";

export type SceneStates = Extract<
  Awaited<ReturnType<typeof createBaseSceneStates>>,
  { error: null }
>;

// export type SceneStates =   Awaited<ReturnType<typeof createBaseSceneStates>>,

export type SceneStatesWithHelper = ReturnType<
  typeof createSceneStatesWithHelper
>;
