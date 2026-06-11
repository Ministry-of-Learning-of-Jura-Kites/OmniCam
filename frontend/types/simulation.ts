export type SimulationState = "idle" | "running" | "finished" | "paused";
export interface Simulation {
  areas: SimulationArea[];
  populationGroups: PopulationGroup[];
  routes: SimulationRoute[];
}

export interface SimulationArea {
  id: string;
  name: string;
  color: string;
  kind: "start" | "end";
  points: [number, number, number][];
}

export interface PopulationGroup {
  id: string;

  routeId: string;

  height: number;
  speed: number;
  count: number;
}

export interface SimulationRoute {
  id: string;
  name: string;
  startAreaId: string;
  endAreaId: string;

  segments: RouteSegment[];
}

export type RouteSegment = LineSegment | BezierSegment;

export interface LineSegment {
  type: "line";
  points: [number, number, number][];
}

export interface BezierSegment {
  type: "bezier";
  points: [number, number, number][];
}

export interface DraggingWaypoint {
  routeId: string;
  segmentIndex: number;
  pointIndex: number; // 0=start, 1=control(bezier), 2=end
}
