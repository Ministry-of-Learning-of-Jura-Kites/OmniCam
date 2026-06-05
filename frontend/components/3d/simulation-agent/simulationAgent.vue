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

import { useGLTF } from "@tresjs/cientos";
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
  Clock,
} from "three";

import * as SkeletonUtils from "three/examples/jsm/utils/SkeletonUtils.js";

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

  console.log("[SimulationAgent] Native model height:", glbNativeHeight);

  console.log(
    "[SimulationAgent] Animations:",
    sourceAnimations.map((a) => a.name),
  );
}

interface AgentObject {
  id: string;
  groupId: string;

  sceneObject: Object3D;

  mixer: AnimationMixer;

  startPos: Vector3;
  endPos: Vector3;

  speed: number;
}

const agentObjects = shallowRef<AgentObject[]>([]);

const population = computed(
  () => sceneStates.value?.simulation.populationGroups ?? [],
);

const activeAgentByGroup = shallowRef<Record<string, Object3D | null>>({});
const progressByGroup = shallowRef<Record<string, number>>({});

const clock = new Clock();

let rafId = 0;

function checkSimulationFinished() {
  const groups = population.value;

  const allFinished = groups.every(
    (group) => (progressByGroup.value[group.id] ?? 0) >= group.count,
  );

  if (allFinished) {
    sceneStates.value!.isSimulationRunning.isRunning = false;
  }
}

function finishAgent(agent: AgentObject) {
  const groupId = agent.groupId;

  // remove current agent
  agentObjects.value = agentObjects.value.filter((x) => x.id !== agent.id);

  // mark progress
  progressByGroup.value[groupId] = (progressByGroup.value[groupId] ?? 0) + 1;

  // allow next spawn
  activeAgentByGroup.value[groupId] = null;

  spawnAgents();

  checkSimulationFinished();
}

function animate() {
  rafId = requestAnimationFrame(animate);

  const delta = clock.getDelta();

  for (const agent of [...agentObjects.value]) {
    // animation clip update
    agent.mixer.update(delta);

    // movement direction
    const direction = agent.endPos.clone().sub(agent.sceneObject.position);

    const distance = direction.length();

    // destination reached
    if (distance <= 0.1) {
      finishAgent(agent);
      continue;
    }

    direction.normalize();

    // move forward
    agent.sceneObject.position.add(
      direction.multiplyScalar(agent.speed * delta),
    );
  }
}
onMounted(() => {
  animate();
});

onUnmounted(() => {
  cancelAnimationFrame(rafId);
});

function getFaceCenterWorld(face: ProcessedCoverageFace): Vector3 {
  const [p0, p1, _p2, p3] = face.points;

  return new Vector3(
    (p0[0] + p1[0] + p3[0]) / 3,
    (p0[1] + p1[1] + p3[1]) / 3,
    (p0[2] + p1[2] + p3[2]) / 3,
  ).applyMatrix4(new Matrix4());
}

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
        mat.alphaTest = 0;
        mat.depthWrite = false;
      } else {
        mat.transparent = false;
        mat.alphaTest = 0.5;
        mat.depthWrite = true;
      }

      mat.side = DoubleSide;

      if (mat.map) {
        mat.map.colorSpace = SRGBColorSpace;
      }

      if (mat.emissiveMap) {
        mat.emissiveMap.colorSpace = SRGBColorSpace;
      }

      mat.needsUpdate = true;
    }

    const matName = Array.isArray(obj.material)
      ? obj.material[0]?.name
      : obj.material?.name;

    obj.renderOrder = RENDER_ORDER[matName] ?? 0;
  });
}

function spawnAgents() {
  const groups = population.value;

  const faces = sceneStates.value!.facesManagement.faces as Record<
    string,
    ProcessedCoverageFace
  >;

  for (const group of groups) {
    if (!group) continue;

    const progress = progressByGroup.value[group.id] ?? 0;
    const active = activeAgentByGroup.value[group.id];

    // already walking
    if (active) continue;

    // completed all agents
    if (progress >= group.count) continue;

    const startEntry = Object.entries(faces).find(
      ([id, face]) =>
        face.type === "simulation" &&
        face.kind === "start" &&
        id === group.startAreaId,
    );

    const endEntry = Object.entries(faces).find(
      ([id, face]) =>
        face.type === "simulation" &&
        face.kind === "end" &&
        id === group.endAreaId,
    );

    if (!startEntry || !endEntry) continue;

    const startPos = getFaceCenterWorld(startEntry[1]);
    const endPos = getFaceCenterWorld(endEntry[1]);

    const scale = group.height / glbNativeHeight;

    const model = SkeletonUtils.clone(sourceScene);

    fixMaterials(model);

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

    // rotate toward destination
    root.lookAt(endPos);

    const agent: AgentObject = {
      id: `${group.id}__${progress}`,
      groupId: group.id,

      sceneObject: root,

      mixer,

      startPos: startPos.clone(),
      endPos: endPos.clone(),

      speed: 1.4,
    };

    agentObjects.value.push(agent);

    activeAgentByGroup.value[group.id] = root;
  }
}

watchEffect(() => {
  const running = sceneStates.value?.isSimulationRunning.isRunning;

  if (!running) {
    agentObjects.value = [];
    activeAgentByGroup.value = {};
    progressByGroup.value = {};
    return;
  }

  spawnAgents();
});
</script>

<template>
  <primitive
    v-for="agent in agentObjects"
    :key="agent.id"
    :object="agent.sceneObject"
  />
</template>
