export interface Simulation {
  areas: SimulationArea[];
  populationGroups: PopulationGroup[];
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

  startAreaId: string;
  endAreaId: string;

  height: number;
  speed: number;
  count: number;
}
