<script setup lang="ts">
import {
  markRaw,
  shallowRef,
  inject,
  watchEffect,
  computed,
  onMounted,
  onUnmounted,
} from "vue";

import { SCENE_STATES_KEY } from "~/constants/state-keys";
import humanGlbUrl from "~/assets/models/human-fixed.glb?url";

import type { ProcessedCoverageFace } from "../scene-states-provider/create-scene-states";
import type { Object3D } from "three";

import {
  Vector3,
  Box3,
  Group,
  Matrix4,
  SRGBColorSpace,
  DoubleSide,
  AnimationMixer,
  Mesh,
  Timer,
  QuadraticBezierCurve3,
} from "three";

import * as SkeletonUtils from "three/examples/jsm/utils/SkeletonUtils.js";
import { useGLTF } from "@tresjs/cientos";
import type { SimulationRoute } from "~/types/simulation";

const sceneStates = inject(SCENE_STATES_KEY)!;

const gltfLoader = await useGLTF(humanGlbUrl);
const gltf = await gltfLoader.execute();

const sourceScene = gltf.scene;
const sourceAnimations = gltf.animations;

let glbNativeHeight = 1;

{
  const box = new Box3().setFromObject(sourceScene);
  const size = new Vector3();
  box.getSize(size);
  glbNativeHeight = size.y || 1;
}

interface AgentObject {
  id: string;
  groupId: string;
  sceneObject: Object3D;
  mixer: AnimationMixer;
  startPos: Vector3;
  endPos: Vector3;
  speed: number;
  path: Vector3[];
  currentWaypoint: number;
}

const agentObjects = shallowRef<AgentObject[]>([]);
const population = computed(
  () => sceneStates.value?.simulation.populationGroups ?? [],
);
const activeAgentByGroup = shallowRef<Record<string, Object3D | null>>({});
const progressByGroup = shallowRef<Record<string, number>>({});
const clock = new Timer();
let rafId = 0;

const NEEDS_TRANSPARENCY = new Set([
  "Hairmat",
  "Beardmat",
  "Moustachemat",
  "Eyewearmat",
]);

const RENDER_ORDER: Record<string, number> = {
  Bodymat: 0,
  Bodymat_0: 0,
  Bodymat_1: 0,
  Bottommat: 1,
  Topmat: 1,
  Shoesmat: 1,
  Eyewearmat: 2,
  Beardmat: 3,
  Moustachemat: 3,
  Hairmat: 4,
};

function fixMaterials(root: Object3D): void {
  root.traverse((obj: Object3D) => {
    if (!(obj instanceof Mesh)) return;
    obj.castShadow = true;
    obj.receiveShadow = true;
    const mats = Array.isArray(obj.material) ? obj.material : [obj.material];
    for (const mat of mats) {
      if (!mat) continue;
      const needsBlend = NEEDS_TRANSPARENCY.has(mat.name);
      if (needsBlend) {
        mat.transparent = true;
        mat.depthWrite = false;
      } else {
        mat.transparent = false;
        mat.depthWrite = true;
      }
      mat.side = DoubleSide;
      if (mat.map) mat.map.colorSpace = SRGBColorSpace;
      mat.needsUpdate = true;
    }
    const matName = Array.isArray(obj.material)
      ? obj.material[0]?.name
      : obj.material?.name;
    obj.renderOrder = RENDER_ORDER[matName] ?? 0;
  });
}

function getFaceCenterWorld(face: ProcessedCoverageFace): Vector3 {
  const [p0, p1, , p3] = face.points;
  return new Vector3(
    (p0[0] + p1[0] + p3[0]) / 3,
    (p0[1] + p1[1] + p3[1]) / 3,
    (p0[2] + p1[2] + p3[2]) / 3,
  ).applyMatrix4(new Matrix4());
}

function buildPathFromRoute(
  route: SimulationRoute | undefined,
  startPos: Vector3,
  endPos: Vector3,
): Vector3[] {
  if (!route || route.segments.length === 0) {
    return [startPos.clone(), endPos.clone()];
  }

  const path: Vector3[] = [];
  path.push(startPos.clone());

  for (const seg of route.segments) {
    if (seg.type === "line") {
      const end = seg.points[seg.points.length - 1];
      if (end) {
        path.push(new Vector3(end[0], end[1], end[2]));
      }
    } else if (seg.type === "bezier" && seg.points.length === 3) {
      const [a, ctrl, b] = seg.points;
      if (a && ctrl && b) {
        const curve = new QuadraticBezierCurve3(
          new Vector3(a[0], a[1], a[2]),
          new Vector3(ctrl[0], ctrl[1], ctrl[2]),
          new Vector3(b[0], b[1], b[2]),
        );
        const sampled = curve.getPoints(16);
        for (let i = 1; i < sampled.length; i++) {
          path.push(sampled[i]!);
        }
      }
    }
  }

  path.push(endPos.clone());
  return path;
}

function checkSimulationFinished() {
  const allFinished = population.value.every(
    (group) => (progressByGroup.value[group.id] ?? 0) >= group.count,
  );
  if (allFinished) {
    sceneStates.value!.simulationState.value = "finished";
  }
}

function finishAgent(agent: AgentObject) {
  agentObjects.value = agentObjects.value.filter((x) => x.id !== agent.id);
  progressByGroup.value[agent.groupId] =
    (progressByGroup.value[agent.groupId] ?? 0) + 1;
  activeAgentByGroup.value[agent.groupId] = null;
  spawnAgents();
  checkSimulationFinished();
}

function lerpAngle(a: number, b: number, t: number) {
  const diff = Math.atan2(Math.sin(b - a), Math.cos(b - a));
  return a + diff * t;
}

function animate() {
  rafId = requestAnimationFrame(animate);
  const state = sceneStates.value?.simulationState.value;

  clock.update();

  if (state !== "running") {
    clock.getDelta();
    return;
  }
  const delta = clock.getDelta();

  for (const agent of [...agentObjects.value]) {
    agent.mixer.update(delta);

    const target = agent.path[agent.currentWaypoint];

    if (!target) {
      finishAgent(agent);
      continue;
    }

    const direction = target.clone().sub(agent.sceneObject.position);
    const distance = direction.length();

    if (distance < 0.15) {
      agent.currentWaypoint++;
      continue;
    }

    direction.normalize();

    agent.sceneObject.position.add(
      direction.clone().multiplyScalar(agent.speed * delta),
    );

    const targetAngle = Math.atan2(direction.x, direction.z);
    agent.sceneObject.rotation.y = lerpAngle(
      agent.sceneObject.rotation.y,
      targetAngle,
      0.15,
    );
  }
}

onMounted(() => animate());
onUnmounted(() => cancelAnimationFrame(rafId));

function spawnAgents() {
  const groups = population.value;
  const faces = sceneStates.value!.facesManagement.faces as Record<
    string,
    ProcessedCoverageFace
  >;

  for (const group of groups) {
    const progress = progressByGroup.value[group.id] ?? 0;
    if (progress >= group.count) continue;
    if (activeAgentByGroup.value[group.id]) continue;

    const route = sceneStates.value?.simulation.routes.find(
      (r) => r.id === group.routeId,
    );
    if (!route) continue;

    const startFaceEntry = Object.entries(faces).find(
      ([id, f]) =>
        f.type === "simulation" &&
        f.kind === "start" &&
        id === route.startAreaId,
    );
    const endFaceEntry = Object.entries(faces).find(
      ([id, f]) =>
        f.type === "simulation" && f.kind === "end" && id === route.endAreaId,
    );
    if (!startFaceEntry || !endFaceEntry) continue;

    const startPos = getFaceCenterWorld(startFaceEntry[1]);
    const endPos = getFaceCenterWorld(endFaceEntry[1]);

    const path = buildPathFromRoute(route, startPos, endPos);

    const model = SkeletonUtils.clone(sourceScene);
    fixMaterials(model);

    const scale = group.height / glbNativeHeight;
    model.scale.setScalar(scale);

    const mixer = new AnimationMixer(model);
    if (sourceAnimations.length > 0) {
      const action = mixer.clipAction(sourceAnimations[0]!);
      action.reset();
      action.play();
    }

    const root = markRaw(new Group());
    root.add(model);
    root.position.copy(startPos);

    const agent: AgentObject = {
      id: `${group.id}_${progress}`,
      groupId: group.id,
      sceneObject: root,
      mixer,
      startPos,
      endPos,
      speed: group.speed ?? 1.4,
      path,
      currentWaypoint: 1,
    };

    agentObjects.value.push(agent);
    activeAgentByGroup.value[group.id] = root;
  }
}

watchEffect(() => {
  const state = sceneStates.value?.simulationState.value;

  switch (state) {
    case "idle":
      agentObjects.value = [];
      activeAgentByGroup.value = {};
      progressByGroup.value = {};
      break;

    case "running":
      spawnAgents();
      break;

    case "finished":
      break;
  }
});
</script>

<template>
  <primitive
    v-for="agent in agentObjects"
    :key="agent.id"
    :object="agent.sceneObject"
  />
</template>
