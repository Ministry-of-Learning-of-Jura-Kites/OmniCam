import type { SceneStatesWithHelper } from "~/types/scene-states";

export function usePromptUnsaved(sceneStates: SceneStatesWithHelper) {
  const handleBeforeUnload = (event: BeforeUnloadEvent) => {
    if (sceneStates.markedForCheck.value) {
      const message =
        "You have unsaved camera changes. Are you sure you want to leave?";
      event.preventDefault();
      event.returnValue = true;
      return message;
    }
  };
  onMounted(() => {
    window.addEventListener("beforeunload", handleBeforeUnload);
  });
  onUnmounted(() => {
    window.removeEventListener("beforeunload", handleBeforeUnload);
  });
  onBeforeRouteLeave((to, from, next) => {
    if (sceneStates.markedForCheck.value) {
      const answer = window.confirm(
        "You have unsaved camera changes. Are you sure you want to leave?",
      );
      if (!answer) {
        next(false);
        return;
      }
    }
    next();
  });
}
