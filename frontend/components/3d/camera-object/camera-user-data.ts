import * as THREE from "three";
import type { IUserData } from "~/components/3d/scene-3d/obj-event-handler";
import { ARROW_CONFIG } from "~/constants";
import type { TresContext } from "@tresjs/core";

export class CameraUserData implements IUserData {
  type: "x" | "y" | "z";
  obj: THREE.Mesh;
  context: TresContext;

  isDragging = false;

  constructor(type: string, obj: THREE.Mesh, context: TresContext) {
    this.type = type as "x" | "y" | "z";
    this.obj = obj;
    this.context = context;
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

    const camera = this.context.camera.value!;

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

    if (this.obj.position[this.type] != null) {
      this.obj.position[this.type] +=
        (projectedVector.x * event.movementX -
          projectedVector.y * event.movementY) *
        ARROW_CONFIG.DRAGGING_SENTIVITY;
    }
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
