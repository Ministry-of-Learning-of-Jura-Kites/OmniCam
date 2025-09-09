import type { TresContext } from "@tresjs/core";
import type * as THREE from "three";
import type { IUserData } from "~/types/obj-3d-user-data";

export class RotatingUserData implements IUserData {
  type: "x" | "y" | "z";
  obj: THREE.Mesh;
  context: TresContext;

  isDragging = false;

  downAngle: number | undefined = undefined;

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

  handlePointerDownEvent = (event: PointerEvent) => {
    this.isDragging = true;
    document.addEventListener("pointermove", this.handlePointerMoveEvent);
    document.addEventListener("pointerup", this.handlePointerUpEvent);
    const objNdc = this.obj.position.clone();
    objNdc.project(this.context.camera.value!);

    let downAngle = Math.atan(
      (event.clientY - objNdc.y) / (event.clientX - objNdc.x),
    );
    if (!Number.isFinite(downAngle)) {
      downAngle = 0;
    }
    this.downAngle = downAngle;
  };

  handlePointerMoveEvent = (event: PointerEvent) => {
    event.clientX;
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
