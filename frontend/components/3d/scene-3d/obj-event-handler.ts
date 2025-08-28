import type * as THREE from "three";

export interface IUserData {
  type: string;
  obj: THREE.Mesh;
  handleEvent: (this: IUserData, eventType: string, event: Event) => void;
}

// const eventMapper = {
//   "ARROW_Z":{
//     "pointerdown":(obj: THREE.Object3DEventMap,userData: UserData,event: PointerEvent)=>{
//       obj.
//       userData.obj.value.position+=
//   }
//   }
// }
