import type { SceneStates } from "~/types/scene-states";

export function updateAspectOnResize(sceneStates: SceneStates) {
  const canvas = sceneStates.tresContext.value?.renderer.domElement;

  const origWidth = canvas?.clientWidth ?? 0;
  const origHeight = canvas?.clientHeight ?? 0;

  const aspect =
    sceneStates.currentCam.value.aspectWidth /
    sceneStates.currentCam.value.aspectHeight;

  let width: number;
  let height: number;

  if (aspect == 0) {
    width = origWidth;
    height = origHeight;
  } else if (origWidth > origHeight * aspect) {
    width = aspect * origHeight;
    height = origHeight;
    sceneStates.aspectMarginType.value = "vertical";
  } else {
    width = origWidth;
    height = origWidth / aspect;
    sceneStates.aspectMarginType.value = "horizontal";
  }

  sceneStates.screenSize.width = width;
  sceneStates.screenSize.height = height;

  const canvasSize = {
    width: canvas?.clientWidth,
    height: canvas?.clientHeight,
  };

  if ((canvasSize?.height ?? 0) == (height ?? 0)) {
    sceneStates.aspectMargin.value = {
      width: ((canvasSize?.width ?? 0) - (width ?? 0)) / 2 + "px",
      height: "100%",
    };
  } else {
    sceneStates.aspectMargin.value = {
      width: "100%",
      height: ((canvasSize?.height ?? 0) - (height ?? 0)) / 2 + "px",
    };
  }
}

export function useAspectRatio(sceneStates: SceneStates) {
  const handleResize = () => {
    updateAspectOnResize(sceneStates);
  };

  onMounted(() => {
    window.addEventListener("resize", handleResize);
  });

  onUnmounted(() => {
    window.removeEventListener("resize", handleResize);
  });

  watch(sceneStates.currentCamId, (_) => {
    handleResize();
  });

  return { handleResize };
}
