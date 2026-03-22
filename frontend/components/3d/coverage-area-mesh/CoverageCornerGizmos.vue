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
type Axis = "x" | "y" | "z";

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
  return new MeshBasicMaterial({
    color,
    transparent: true,
    opacity: 0.95,
    depthWrite: false,
    depthTest: false,
  });
}

function makeArrowGeometry() {
  const g = new ConeGeometry(arrowRad, arrowLen, 10);
  // ทำให้ฐานอยู่ที่ origin แล้วปลายยื่นออกไปตามแกน local +Y
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

function createAxisArrow(
  axis: Axis,
  dir: 1 | -1,
  cornerIndex: number,
): Obj3DWithUserData {
  let material: MeshBasicMaterial;

  if (axis === "x") material = matX;
  else if (axis === "y") material = matY;
  else material = matZ;

  const arrow = new Mesh(
    geomBase.clone(),
    material,
  ) as unknown as Obj3DWithUserData;

  // geometry เดิมชี้ local +Y
  // หมุนให้ชี้ตามแกนที่ต้องการ
  if (axis === "x") {
    arrow.rotateZ(dir === 1 ? -Math.PI / 2 : Math.PI / 2);
  } else if (axis === "y") {
    if (dir === -1) {
      arrow.rotateZ(Math.PI);
    }
  } else if (axis === "z") {
    arrow.rotateX(dir === 1 ? Math.PI / 2 : -Math.PI / 2);
  }

  // ใช้ axis เดิมได้เลย เพราะ logic drag เป็นเส้นแกนเต็มอยู่แล้ว
  arrow.userData = new CornerTranslateUserData(
    axis,
    props.faceId,
    cornerIndex,
    sceneStates,
    context,
    props.yOffset,
  );

  return arrow;
}

// 1 จุด = 1 dot + 6 arrows
for (let ci = 0; ci < 4; ci++) {
  const dot = new Mesh(dotGeom, dotMat);
  dots.push(dot);
  group.add(dot);

  const axPos = createAxisArrow("x", 1, ci);
  const axNeg = createAxisArrow("x", -1, ci);
  const ayPos = createAxisArrow("y", 1, ci);
  const ayNeg = createAxisArrow("y", -1, ci);
  const azPos = createAxisArrow("z", 1, ci);
  const azNeg = createAxisArrow("z", -1, ci);

  arrows.push(axPos, axNeg, ayPos, ayNeg, azPos, azNeg);

  group.add(axPos);
  group.add(axNeg);
  group.add(ayPos);
  group.add(ayNeg);
  group.add(azPos);
  group.add(azNeg);
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

    // 1 corner มี 6 arrows
    for (let k = 0; k < 6; k++) {
      arrows[ci * 6 + k]!.position.copy(pos);
    }
  }
});

onBeforeUnmount(() => {
  arrows.forEach((m) => sceneStates.draggableObjects.delete(m));

  dots.forEach((d) => group.remove(d));
  arrows.forEach((a) => group.remove(a));

  dotGeom.dispose();
  geomBase.dispose();
  arrows.forEach((a) =>
    ((a as unknown as Mesh).geometry as BufferGeometry)?.dispose?.(),
  );

  dotMat.dispose();
  matX.dispose();
  matY.dispose();
  matZ.dispose();
});
</script>

<template>
  <primitive :object="group" />
</template>
