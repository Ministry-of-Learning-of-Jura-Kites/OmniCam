import type { TresContext } from "@tresjs/core";
import type { ICamera } from "~/types/camera";
import type { IUserData } from "~/types/obj-3d-user-data";
import { Quaternion, Vector3 } from "three";
import { getAxisVector } from "~/lib/three";

export const ROTATING_TYPE = "rotation";

export class RotatingUserData implements IUserData<ICamera> {
  type: "x" | "y" | "z";
  target: ICamera;
  context: TresContext;

  isDragging = false;

  initialCamQuaternion = new Quaternion();
  downAngle = 0;

  constructor(type: string, obj: ICamera, context: TresContext) {
    this.type = type as "x" | "y" | "z";
    this.target = obj;
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
    this.target.controlling = {
      type: ROTATING_TYPE,
      direction: this.type,
    };
    document.addEventListener("pointermove", this.handlePointerMoveEvent);
    document.addEventListener("pointerup", this.handlePointerUpEvent);

    this.initialCamQuaternion.setFromEuler(this.target.rotation);

    const ele = this.context.renderer.value.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height) * 2 + 1;
    const objNdc = this.target.position
      .clone()
      .project(this.context.camera.value!);
    this.downAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);
  };

  handlePointerMoveEvent = (event: PointerEvent) => {
    const objNdc = this.target.position.clone();
    objNdc.project(this.context.camera.value!);

    const ele = this.context.renderer.value.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height) * 2 + 1;

    const moveAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);

    const delta = moveAngle - this.downAngle!;

    const cameraDirection = new Vector3();

    this.context.camera.value!.getWorldDirection(cameraDirection);

    // Clockwise/Anticlockwise depend on side of object looked from
    const direction = -Math.sign(cameraDirection[this.type]);

    const quaternion = this.initialCamQuaternion.clone();
    const rotation = new Quaternion().setFromAxisAngle(
      getAxisVector(this.type).multiplyScalar(direction),
      delta,
    );
    quaternion.premultiply(rotation);

    this.target.rotation.setFromQuaternion(quaternion);
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    this.target.controlling = undefined;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
