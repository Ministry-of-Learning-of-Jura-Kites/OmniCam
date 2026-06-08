<script setup lang="ts">
import { inject, computed } from "vue";
import { SCENE_STATES_KEY } from "~/constants/state-keys";
import { Vector3, QuadraticBezierCurve3 } from "three";
import type { RouteSegment } from "~/types/simulation";

const sceneStates = inject(SCENE_STATES_KEY)!;

const Y_OFFSET = 0.03;

function p3toV3(p: [number, number, number]): Vector3 {
  return new Vector3(p[0], p[1] + Y_OFFSET, p[2]);
}

function sampleBezier(
  p0: Vector3,
  ctrl: Vector3,
  p2: Vector3,
  steps = 48,
): Vector3[] {
  return new QuadraticBezierCurve3(p0, ctrl, p2).getPoints(steps);
}

function segmentToPoints(seg: RouteSegment): Vector3[] {
  if (seg.type === "line") {
    const [a, b] = seg.points;
    if (!a || !b) return [];
    return [
      p3toV3(a as [number, number, number]),
      p3toV3(b as [number, number, number]),
    ];
  }
  const [a, ctrl, b] = seg.points;
  if (!a || !ctrl || !b) return [];
  return sampleBezier(
    p3toV3(a as [number, number, number]),
    p3toV3(ctrl as [number, number, number]),
    p3toV3(b as [number, number, number]),
  );
}

interface RenderRoute {
  id: string;
  isActive: boolean;
  segments: {
    index: number;
    type: "line" | "bezier";
    linePoints: Vector3[];
    controlPoint: Vector3 | null;
    startPoint: Vector3 | null;
    endPoint: Vector3 | null;
  }[];
}

const renderRoutes = computed<RenderRoute[]>(() => {
  const routes = sceneStates.value?.simulation.routes ?? [];
  const activeId = sceneStates.value?.routeDrawing.activeRouteId ?? null;

  return routes
    .filter((r) => r.segments.length > 0)
    .map((route) => ({
      id: route.id,
      isActive: route.id === activeId,
      segments: route.segments.map((seg, index) => {
        const pts = segmentToPoints(seg);
        let controlPoint: Vector3 | null = null;
        let startPoint: Vector3 | null = null;
        let endPoint: Vector3 | null = null;

        if (seg.type === "bezier" && seg.points.length === 3) {
          startPoint = p3toV3(seg.points[0] as [number, number, number]);
          controlPoint = p3toV3(seg.points[1] as [number, number, number]);
          endPoint = p3toV3(seg.points[2] as [number, number, number]);
        }

        return {
          index,
          type: seg.type,
          linePoints: pts,
          controlPoint,
          startPoint,
          endPoint,
        };
      }),
    }));
});

function toPositionsArray(pts: Vector3[]): number[] {
  return pts.flatMap((p) => [p.x, p.y, p.z]);
}

function toFloat32(arr: number[]): Float32Array {
  return new Float32Array(arr);
}
</script>

<template>
  <template v-for="route in renderRoutes" :key="route.id">
    <template v-for="seg in route.segments" :key="`${route.id}-${seg.index}`">
      <!-- ── Line / Bezier curve ─────────────────────────────── -->
      <TresLine v-if="seg.linePoints.length >= 2" :render-order="50">
        <TresBufferGeometry
          :position="[toFloat32(toPositionsArray(seg.linePoints)), 3]"
        />
        <TresLineBasicMaterial
          :color="route.isActive ? '#facc15' : '#60a5fa'"
          :depth-test="false"
        />
      </TresLine>

      <!-- ── Start waypoint dot — always draggable ─────────── -->
      <TresMesh
        v-if="seg.linePoints[0]"
        :position="[
          seg.linePoints[0].x,
          seg.linePoints[0].y,
          seg.linePoints[0].z,
        ]"
        :render-order="51"
        :user-data="{
          type: 'waypoint',
          routeId: route.id,
          segmentIndex: seg.index,
          pointIndex: 0,
        }"
      >
        <TresSphereGeometry :args="[0.05, 12, 12]" />
        <TresMeshBasicMaterial color="#ffffff" :depth-test="false" />
      </TresMesh>

      <!-- ── End waypoint dot — always draggable ───────────── -->
      <TresMesh
        v-if="seg.linePoints[seg.linePoints.length - 1]"
        :position="[
          seg.linePoints[seg.linePoints.length - 1]!.x,
          seg.linePoints[seg.linePoints.length - 1]!.y,
          seg.linePoints[seg.linePoints.length - 1]!.z,
        ]"
        :render-order="51"
        :user-data="{
          type: 'waypoint',
          routeId: route.id,
          segmentIndex: seg.index,
          // end point is last index in points array
          pointIndex: seg.type === 'bezier' ? 2 : 1,
        }"
      >
        <TresSphereGeometry :args="[0.05, 12, 12]" />
        <TresMeshBasicMaterial color="#ffffff" :depth-test="false" />
      </TresMesh>

      <!-- ── Bezier control point + handles (active route only) ── -->
      <template
        v-if="seg.type === 'bezier' && route.isActive && seg.controlPoint"
      >
        <TresLine v-if="seg.startPoint" :render-order="50">
          <TresBufferGeometry
            :position="[
              toFloat32([
                seg.startPoint.x,
                seg.startPoint.y,
                seg.startPoint.z,
                seg.controlPoint!.x,
                seg.controlPoint!.y,
                seg.controlPoint!.z,
              ]),
              3,
            ]"
          />
          <TresLineBasicMaterial color="#f97316" :depth-test="false" />
        </TresLine>

        <TresLine v-if="seg.endPoint" :render-order="50">
          <TresBufferGeometry
            :position="[
              toFloat32([
                seg.controlPoint!.x,
                seg.controlPoint!.y,
                seg.controlPoint!.z,
                seg.endPoint.x,
                seg.endPoint.y,
                seg.endPoint.z,
              ]),
              3,
            ]"
          />
          <TresLineBasicMaterial color="#f97316" :depth-test="false" />
        </TresLine>

        <!-- Control point handle — always draggable -->
        <TresMesh
          :position="[
            seg.controlPoint!.x,
            seg.controlPoint!.y,
            seg.controlPoint!.z,
          ]"
          :render-order="52"
          :user-data="{
            type: 'waypoint',
            routeId: route.id,
            segmentIndex: seg.index,
            pointIndex: 1,
          }"
        >
          <TresSphereGeometry :args="[0.06, 12, 12]" />
          <TresMeshBasicMaterial color="#f97316" :depth-test="false" />
        </TresMesh>
      </template>
    </template>
  </template>
</template>
