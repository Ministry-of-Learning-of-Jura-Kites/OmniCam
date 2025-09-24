import type * as THREE from "three";
import type { ICamera } from "./camera";

export type Obj3DWithUserData = THREE.Object3D & { userData: IUserData };

export interface IUserData {
  type: string;
  // obj: ModelRef<THREE.Mesh>;
  cam: ICamera;
  handleEvent: (this: IUserData, eventType: string, event: Event) => void;
}
