<script setup lang="ts">
import { inject, watchEffect, onBeforeUnmount } from "vue";
import { useTresContext } from "@tresjs/core";
import {
  ConeGeometry,
  Group,
  Mesh,
  MeshBasicMaterial,
  SphereGeometry,
  Vector3,
} from "three";
import type { BufferGeometry } from "three";
import { SCENE_STATES_KEY } from "@/constants/state-keys";
import type { Obj3DWithUserData } from "~/types/obj-3d-user-data";
import { CornerTranslateUserData } from "./corner-translate-handle";

type Point3 = [number, number, number];

const props = withDefaults(
  defineProps<{
    faceId: string;
    points: Point3[];
    size?: number;
    yOffset?: number;
    visible?: boolean;
  }>(),
  {
    size: 0.14,
    yOffset: 0.01,
    visible: true,
  },
);

const sceneStates = inject(SCENE_STATES_KEY)!;
const context = useTresContext();

const group = new Group();
group.visible = props.visible;

const arrowLen = props.size;
const arrowRad = props.size * 0.12;

function makeArrowMaterial(color: string) {
  const m = new MeshBasicMaterial({
    color,
    transparent: true,
    opacity: 0.95,
    depthWrite: false,
    depthTest: false,
  });
  return m;
}

function makeArrowGeometry() {
  const g = new ConeGeometry(arrowRad, arrowLen, 10);
  g.translate(0, arrowLen / 2, 0);
  return g;
}

const geomBase = makeArrowGeometry();
const matX = makeArrowMaterial("green");
const matY = makeArrowMaterial("red");
const matZ = makeArrowMaterial("blue");

const dotGeom = new SphereGeometry(props.size * 0.12, 10, 10);
const dotMat = new MeshBasicMaterial({
  color: "white",
  transparent: true,
  opacity: 0.85,
  depthWrite: false,
  depthTest: false,
});

const arrows: Obj3DWithUserData[] = [];
const dots: Mesh[] = [];

for (let ci = 0; ci < 4; ci++) {
  const dot = new Mesh(dotGeom, dotMat);
  dots.push(dot);
  group.add(dot);

  const ax = new Mesh(geomBase.clone(), matX) as unknown as Obj3DWithUserData;
  ax.rotateZ(-Math.PI / 2);
  ax.userData = new CornerTranslateUserData(
    "x",
    props.faceId,
    ci,
    sceneStates,
    context,
    props.yOffset,
  );
  arrows.push(ax);
  group.add(ax);

  const ay = new Mesh(geomBase.clone(), matY) as unknown as Obj3DWithUserData;
  ay.userData = new CornerTranslateUserData(
    "y",
    props.faceId,
    ci,
    sceneStates,
    context,
    props.yOffset,
  );
  arrows.push(ay);
  group.add(ay);

  const az = new Mesh(geomBase.clone(), matZ) as unknown as Obj3DWithUserData;
  az.rotateX(Math.PI / 2);
  az.userData = new CornerTranslateUserData(
    "z",
    props.faceId,
    ci,
    sceneStates,
    context,
    props.yOffset,
  );
  arrows.push(az);
  group.add(az);
}

// register draggable
arrows.forEach((m) => sceneStates.draggableObjects.add(m));

watchEffect(() => {
  group.visible = !!props.visible;

  const pts = props.points ?? [];
  if (pts.length !== 4) {
    group.visible = false;
    return;
  }

  for (let ci = 0; ci < 4; ci++) {
    const p = pts[ci]!;
    const pos = new Vector3(p[0], p[1] + props.yOffset, p[2]);

    dots[ci]!.position.copy(pos);

    arrows[ci * 3 + 0]!.position.copy(pos);
    arrows[ci * 3 + 1]!.position.copy(pos);
    arrows[ci * 3 + 2]!.position.copy(pos);
  }
});

onBeforeUnmount(() => {
  arrows.forEach((m) => sceneStates.draggableObjects.delete(m));

  dots.forEach((d) => group.remove(d));
  arrows.forEach((a) => group.remove(a));

  // dispose geometries
  dotGeom.dispose();
  geomBase.dispose();
  arrows.forEach((a) =>
    ((a as unknown as Mesh).geometry as BufferGeometry)?.dispose?.(),
  );

  // dispose materials (shared)
  dotMat.dispose();
  matX.dispose();
  matY.dispose();
  matZ.dispose();
});
</script>

<template>
  <primitive :object="group" />
</template>
