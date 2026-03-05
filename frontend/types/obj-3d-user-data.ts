import type { Object3D } from "three";
import type { ICamera } from "./camera";

export type Obj3DWithUserData = Object3D & { userData: IUserData };

export interface IUserData {
  type: string;
  // obj: ModelRef<THREE.Mesh>;
  cam: ICamera;
  handleEvent: (this: IUserData, eventType: string, event: Event) => void;
}
