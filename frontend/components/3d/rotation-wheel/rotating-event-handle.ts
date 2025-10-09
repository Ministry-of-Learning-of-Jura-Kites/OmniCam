import type { TresContext } from "@tresjs/core";
import type { ICamera } from "~/types/camera";
import type { IUserData } from "~/types/obj-3d-user-data";

export const ROTATING_TYPE = "rotation";

export class RotatingUserData implements IUserData {
  type: "x" | "y" | "z";
  cam: ICamera;
  context: TresContext;

  isDragging = false;

  downAngle: number | undefined = undefined;

  constructor(type: string, obj: ICamera, context: TresContext) {
    this.type = type as "x" | "y" | "z";
    this.cam = obj;
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
    this.cam.controlling = {
      type: ROTATING_TYPE,
      direction: this.type,
    };
    document.addEventListener("pointermove", this.handlePointerMoveEvent);
    document.addEventListener("pointerup", this.handlePointerUpEvent);

    const objNdc = this.cam.position.clone();
    objNdc.project(this.context.camera.value!);

    const ele = this.context.renderer.value.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height!) * 2 + 1;

    const downAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);
    this.downAngle = downAngle;
  };

  handlePointerMoveEvent = (event: PointerEvent) => {
    const objNdc = this.cam.position.clone();
    objNdc.project(this.context.camera.value!);

    const ele = this.context.renderer.value.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width!) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height!) * 2 + 1;

    const moveAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);
    this.cam.rotation[this.type] = moveAngle + this.downAngle!;
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    this.cam.controlling = undefined;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
