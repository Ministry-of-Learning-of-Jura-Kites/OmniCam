<script setup lang="ts">
import {
  Mesh,
  CylinderGeometry,
  MeshBasicMaterial,
  Group,
  Scene,
  OrthographicCamera,
  Vector3,
  CircleGeometry,
  DoubleSide,
  Euler,
} from "three";
import { AXIS_COLOR, AXIS_GIZMO } from "~/constants";
import { SCENE_STATES_KEY } from "~/constants/state-keys";
import type { Font } from "three/examples/jsm/loaders/FontLoader.js";
import { FontLoader } from "three/examples/jsm/loaders/FontLoader.js";
import { TextGeometry } from "three/addons/geometries/TextGeometry.js";
import { debouncedWatch } from "@vueuse/core";

const d = 1.5;

const sceneStates = inject(SCENE_STATES_KEY);
const axisScene = new Scene();
const axisCam = new OrthographicCamera(-d, d, d, -d, 0.01, 1000);

const cylinder = new CylinderGeometry(
  AXIS_GIZMO.RADIUS,
  AXIS_GIZMO.RADIUS,
  AXIS_GIZMO.LENGTH,
  8,
);

const axes = new Group();
const labels = new Group();

axisScene.add(labels);
axisScene.add(axes);
axisScene.background = null;

const disposables: { dispose: () => void }[] = [];

const loader = new FontLoader();
// Public CDN font
const fontUrl =
  "https://threejs.org/examples/fonts/helvetiker_regular.typeface.json";

loader.load(fontUrl, (font) => {
  setupLabels(font);
});

const textMat = new MeshBasicMaterial({
  color: "white",
  side: DoubleSide,
  depthTest: false,
});
disposables.push(textMat);

function setupLabels(font: Font) {
  for (const axis of ["x", "y", "z"] as const) {
    const color = AXIS_COLOR[axis];
    const material = new MeshBasicMaterial({ color, side: DoubleSide });
    disposables.push(material);

    const cylinderMesh = new Mesh(cylinder, material);
    cylinderMesh.renderOrder = 3;
    const offset = AXIS_GIZMO.LENGTH / 2;

    axes.add(cylinderMesh);

    const textGeo = new TextGeometry(axis, {
      font: font,
      size: 0.2,
    });
    textGeo.center();
    disposables.push(textGeo);

    const textMesh = new Mesh(textGeo, textMat);
    textMesh.renderOrder = 2;

    const circleGeo = new CircleGeometry(AXIS_GIZMO.LABEL_RADIUS, 32);
    disposables.push(circleGeo);

    const circleMesh = new Mesh(circleGeo, material);
    circleMesh.renderOrder = 1;

    const billboard = new Group();
    billboard.add(circleMesh);
    billboard.add(textMesh);

    switch (axis) {
      case "x":
        cylinderMesh.rotateZ(Math.PI / 2);
        cylinderMesh.position.setX(offset);
        billboard.position.setX(AXIS_GIZMO.LENGTH + AXIS_GIZMO.LABEL_RADIUS);
        break;
      case "y":
        cylinderMesh.position.setY(offset);
        billboard.position.setY(AXIS_GIZMO.LENGTH + AXIS_GIZMO.LABEL_RADIUS);
        break;
      case "z":
        cylinderMesh.rotateX(Math.PI / 2);
        cylinderMesh.position.setZ(offset);
        billboard.position.setZ(AXIS_GIZMO.LENGTH + AXIS_GIZMO.LABEL_RADIUS);
        break;
      default:
        break;
    }

    labels.add(billboard);
  }
}

const eul = new Euler();
const forward = new Vector3(0, 0, 1);

debouncedWatch(
  () => [
    sceneStates?.value?.currentCam.value.rotation.x,
    sceneStates?.value?.currentCam.value.rotation.y,
    sceneStates?.value?.currentCam.value.rotation.z,
  ],
  ([x, y, z]) => {
    if (x == undefined || y == undefined || z == undefined) {
      return;
    }
    forward.set(0, 0, 1);
    eul.set(x, y, z);
    forward.applyEuler(eul);
    axisCam.position.set(forward.x, forward.y, forward.z);
    axisCam.lookAt(0, 0, 0);

    for (const child of labels.children) {
      for (const obj of child.children) {
        // obj.setRotationFromEuler(eul);
        obj.quaternion.copy(axisCam.quaternion);
      }
    }
  },
  { deep: true, debounce: 10, maxWait: 100 },
);

const renderer = sceneStates!.value!.tresContext.value!.renderer;
const onRender = renderer.onRender;
onRender((renderer) => {
  renderer.autoClear = false;
  renderer.clearDepth();
  renderer.setScissorTest(true);
  renderer.setViewport(0, 0, AXIS_GIZMO.GIZMO_WIDTH, AXIS_GIZMO.GIZMO_WIDTH);
  renderer.setScissor(0, 0, AXIS_GIZMO.GIZMO_WIDTH, AXIS_GIZMO.GIZMO_WIDTH);
  renderer.setClearColor(0x000000, 0);
  renderer.render(axisScene, axisCam);

  renderer.setScissorTest(false);
  renderer.setViewport(
    0,
    0,
    renderer.domElement.clientWidth,
    renderer.domElement.clientHeight,
  );
  renderer.autoClear = true;
});

onBeforeUnmount(() => {
  disposables.forEach((obj) => obj.dispose());
  axisScene.clear();
  axes.clear();
  labels.clear();
});
</script>
<template>
  <slot></slot>
</template>
