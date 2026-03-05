import type { SceneStates } from "~/types/scene-states";

function updateAspectOnResize(
  sceneStates: SceneStates,
  origWidth: number,
  origHeight: number,
) {
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
}

function resizeEntriesToSize(entries: ReadonlyArray<ResizeObserverEntry>) {
  return {
    width: entries[0]?.contentRect.width,
    height: entries[0]?.contentRect.height,
  };
}

function updateAspectFromEleGenerator(sceneStates: SceneStates) {
  return () => {
    const parentOfParent =
      sceneStates.tresContext.value?.renderer.domElement?.parentElement
        ?.parentElement;

    const size = parentOfParent?.getBoundingClientRect();
    updateAspectOnResize(sceneStates, size?.width ?? 0, size?.height ?? 0);
  };
}

export function useAspectRatio(sceneStates: SceneStates) {
  if (!import.meta.client) {
    return;
  }

  const observer = new ResizeObserver((entries) => {
    const size = resizeEntriesToSize(entries);
    updateAspectOnResize(sceneStates, size.width ?? 0, size.height ?? 0);
  });

  const updateAspectFromEle = updateAspectFromEleGenerator(sceneStates);

  onMounted(() => {
    watch(
      () => sceneStates.tresContext.value?.renderer.domElement,
      (canvas) => {
        const canvasParent = canvas?.parentElement;
        const parentOfParent = canvasParent?.parentElement;

        if (
          canvas == undefined ||
          canvasParent == undefined ||
          parentOfParent == undefined
        ) {
          return;
        }

        observer.observe(parentOfParent);
      },
      {
        once: true,
      },
    );
  });

  onUnmounted(() => {
    observer.disconnect();
  });

  watch(sceneStates.currentCamId, (_) => {
    updateAspectFromEle();
  });

  return { updateAspectFromEle };
}
