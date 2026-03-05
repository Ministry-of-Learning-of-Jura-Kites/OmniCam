<!-- OmniCam/frontend/components/3d/coverage-area-mesh/CoverageAreaMesh.vue -->
<script setup lang="ts">
import { computed, watchEffect, onUnmounted } from "vue";
import {
  BufferGeometry,
  Float32BufferAttribute,
  Group,
  LineLoop,
  LineBasicMaterial,
  Mesh,
  MeshBasicMaterial,
  SphereGeometry,
  DoubleSide,
} from "three";

type Point3 = [number, number, number];

const props = withDefaults(
  defineProps<{
    faceId?: string;
    points: Point3[];
    color?: string;
    opacity?: number;
    wireframe?: boolean;
    selected?: boolean;
    showCorners?: boolean;
    yOffset?: number;
  }>(),
  {
    faceId: "",
    color: "#22ff88",
    opacity: undefined,
    wireframe: false,
    selected: false,
    showCorners: true,
    yOffset: 0.01,
  },
);

const fillOpacity = computed(() => {
  if (props.opacity != null) return props.opacity;
  return props.selected ? 0.35 : 0.18;
});
const outlineOpacity = computed(() => (props.selected ? 1 : 0.85));
const cornerRadius = computed(() => (props.selected ? 0.07 : 0.05));

// ---------- THREE objects (stable instances) ----------
const group = new Group();

const meshGeometry = new BufferGeometry();
const outlineGeometry = new BufferGeometry();

const meshMaterial = new MeshBasicMaterial({
  transparent: true,
  opacity: 0.2,
  side: DoubleSide,
  depthWrite: false,
  depthTest: false,
  polygonOffset: true,
  polygonOffsetFactor: -2,
  polygonOffsetUnits: -2,
});

const outlineMaterial = new LineBasicMaterial({
  transparent: true,
  opacity: 1,
});
outlineMaterial.depthTest = false;

const mesh = new Mesh(meshGeometry, meshMaterial);
mesh.renderOrder = 999;
mesh.frustumCulled = false;

const outline = new LineLoop(outlineGeometry, outlineMaterial);
outline.renderOrder = 1000;
outline.frustumCulled = false;

group.add(mesh);
group.add(outline);

const cornerGeom = new SphereGeometry(1, 10, 10);
const cornerMaterial = new MeshBasicMaterial({
  transparent: true,
  opacity: 0.85,
  depthWrite: false,
  depthTest: false,
});

const corners = Array.from({ length: 4 }, () => {
  const m = new Mesh(cornerGeom, cornerMaterial);
  m.renderOrder = 1001;
  m.frustumCulled = false;
  group.add(m);
  return m;
});

function applyUserData() {
  const fid = props.faceId;
  mesh.userData = fid
    ? { kind: "coverage-quad", faceId: fid }
    : { kind: "coverage-quad" };
  outline.userData = fid
    ? { kind: "coverage-outline", faceId: fid }
    : { kind: "coverage-outline" };
  corners.forEach((c, i) => {
    c.userData = fid
      ? { kind: "coverage-corner", faceId: fid, cornerIndex: i }
      : { kind: "coverage-corner", cornerIndex: i };
  });
}

// update geometry positions
function setPositions(geom: BufferGeometry, pts: Point3[]) {
  let attr = geom.getAttribute("position") as
    | Float32BufferAttribute
    | undefined;

  if (!attr || (attr.array as Float32Array).length !== 12) {
    attr = new Float32BufferAttribute(new Float32Array(12), 3);
    geom.setAttribute("position", attr);
  }

  const a = attr.array as Float32Array;

  const p0 = pts[0]!;
  const p1 = pts[1]!;
  const p2 = pts[2]!;
  const p3 = pts[3]!;

  a[0] = p0[0];
  a[1] = p0[1];
  a[2] = p0[2];
  a[3] = p1[0];
  a[4] = p1[1];
  a[5] = p1[2];
  a[6] = p2[0];
  a[7] = p2[1];
  a[8] = p2[2];
  a[9] = p3[0];
  a[10] = p3[1];
  a[11] = p3[2];

  attr.needsUpdate = true;
  geom.computeBoundingSphere();
}

let meshIndexSet = false;

watchEffect(() => {
  const pts = props.points ?? [];
  const ok = pts.length === 4;

  group.visible = ok;
  if (!ok) return;

  applyUserData();

  const ep: Point3[] = pts.map((p) => [p[0], p[1] + props.yOffset, p[2]]);

  // mesh
  setPositions(meshGeometry, ep);

  if (!meshIndexSet) {
    meshGeometry.setIndex([0, 1, 2, 0, 2, 3]);
    meshIndexSet = true;
  }
  meshGeometry.computeVertexNormals();

  setPositions(outlineGeometry, ep);

  // materials
  meshMaterial.color.set(props.color);
  meshMaterial.opacity = fillOpacity.value;
  meshMaterial.wireframe = !!props.wireframe;

  outlineMaterial.color.set(props.color);
  outlineMaterial.opacity = outlineOpacity.value;

  // corners
  const show = !!props.showCorners;
  const r = cornerRadius.value;
  corners.forEach((c, i) => {
    c.visible = show;
    const p = ep[i]!;
    c.position.set(p[0], p[1], p[2]);
    c.scale.setScalar(r);
  });

  // corner color
  cornerMaterial.color.set(props.selected ? "#ffffff" : props.color);
  cornerMaterial.opacity = props.selected ? 0.95 : 0.8;
});

onUnmounted(() => {
  // remove children
  group.remove(mesh);
  group.remove(outline);
  corners.forEach((c) => group.remove(c));

  // dispose
  meshGeometry.dispose();
  outlineGeometry.dispose();
  cornerGeom.dispose();

  meshMaterial.dispose();
  outlineMaterial.dispose();
  cornerMaterial.dispose();
});
</script>

<template>
  <primitive :object="group" />
</template>
