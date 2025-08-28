import type * as THREE from "three";

export type ObjWithUserData = THREE.Object3D & { userData: IUserData };

export interface IUserData {
  type: string;
  obj: THREE.Mesh;
  handleEvent: (this: IUserData, eventType: string, event: Event) => void;
}
