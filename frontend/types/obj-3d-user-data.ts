import type { Object3D } from "three";

export type Obj3DWithUserData = Object3D & { userData: IUserData };

export interface IUserData<T = unknown> {
  type: string;
  target: T;
  handleEvent: (eventType: string, event: Event) => void;
}
