import * as THREE from "three";
import type { IUserData } from "~/types/obj-3d-user-data";
import { ARROW_CONFIG } from "~/constants";
import type { SceneStatesWithHelper } from "~/types/scene-states";

export class CameraUserData implements IUserData {
  type: "x" | "y" | "z";
  obj: THREE.Mesh;
  sceneStates: SceneStatesWithHelper;
  camId: string;

  isDragging = false;

  constructor(
    type: string,
    obj: THREE.Mesh,
    sceneStates: SceneStatesWithHelper,
    camId: string,
  ) {
    this.type = type as "x" | "y" | "z";
    this.obj = obj;
    this.sceneStates = sceneStates;
    this.camId = camId;
  }

  handleEvent(eventType: string, event: Event) {
    switch (eventType) {
      case "pointerdown":
        this.handlePointerDownEvent(event as PointerEvent);
        break;
      default:
        break;
    }
  }

  handlePointerDownEvent = (_event: PointerEvent) => {
    this.isDragging = true;
    document.addEventListener("pointermove", this.handlePointerMoveEvent);
    document.addEventListener("pointerup", this.handlePointerUpEvent);
  };

  handlePointerMoveEvent = (event: PointerEvent) => {
    if (!this.isDragging) {
      return;
    }

    const camera = this.sceneStates.tresContext!.value!.camera!;

    const point = this.obj.position.clone();

    let dirVector: THREE.Vector3 | null = null;
    switch (this.type) {
      case "x":
        dirVector = new THREE.Vector3(1, 0, 0);
        break;
      case "y":
        dirVector = new THREE.Vector3(0, 1, 0);
        break;
      case "z":
        dirVector = new THREE.Vector3(0, 0, 1);
        break;
    }

    const vectorEnd = dirVector.add(point);

    const pointNDC = point.clone().project(camera);
    const endNDC = vectorEnd.clone().project(camera);

    const projectedVector = endNDC.sub(pointNDC).normalize();

    // if (this.obj.position[this.type] != null) {
    //   this.obj.position[this.type] +=
    //     (projectedVector.x * event.movementX -
    //       projectedVector.y * event.movementY) *
    //     ARROW_CONFIG.DRAGGING_SENTIVITY;
    // }

    const cam = this.sceneStates!.cameras![this.camId]!;

    const delta =
      (projectedVector.x * event.movementX -
        projectedVector.y * event.movementY) *
      ARROW_CONFIG.DRAGGING_SENTIVITY;

    switch (this.type) {
      case "x":
        cam.position.setX(cam.position.x + delta);
        break;
      case "y":
        cam.position.setY(cam.position.y + delta);
        break;
      case "z":
        cam.position.setZ(cam.position.z + delta);
        break;
    }
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
