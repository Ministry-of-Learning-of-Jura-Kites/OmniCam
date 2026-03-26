import type { TresContext } from "@tresjs/core";
import type { MovableObject } from "~/types/movable";
import type { IUserData } from "~/types/obj-3d-user-data";
import { Quaternion, Vector3 } from "three";
import { getAxisVector } from "~/lib/three";

export const ROTATING_TYPE = "rotation";

export class RotatingUserData implements IUserData<MovableObject> {
  type: "x" | "y" | "z";
  target: MovableObject;
  context: TresContext;

  onUp?: () => void;
  onDown?: () => void;
  onMove:
    | undefined
    | ((direction: "x" | "y" | "z", angleDir: number, delta: number) => void);

  isDragging = false;

  initialCamQuaternion = new Quaternion();
  downAngle = 0;

  constructor(
    type: string,
    obj: MovableObject,
    context: TresContext,
    onUp?: () => void,
    onDown?: () => void,
    onMove?: (
      direction: "x" | "y" | "z",
      directionSign: number,
      delta: number,
    ) => void,
  ) {
    this.type = type as "x" | "y" | "z";
    this.target = obj;
    this.context = context;
    this.onUp = onUp;
    this.onDown = onDown;
    this.onMove = onMove;
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

    const ele = this.context.renderer.instance.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height) * 2 + 1;
    const objNdc = this.target.position
      .clone()
      .project(this.context.camera.activeCamera.value!);
    this.downAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);
    if (this.onDown) {
      this.onDown();
    }
  };

  handlePointerMoveEvent = (event: PointerEvent) => {
    const objNdc = this.target.position.clone();
    objNdc.project(this.context.camera.activeCamera.value!);

    const ele = this.context.renderer.instance.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height) * 2 + 1;

    const moveAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);

    const delta = moveAngle - this.downAngle!;

    const cameraDirection = new Vector3();

    this.context.camera.activeCamera.value!.getWorldDirection(cameraDirection);

    // Clockwise/Anticlockwise depend on side of object looked from
    const direction = -Math.sign(cameraDirection[this.type]);

    const quaternion = this.initialCamQuaternion.clone();
    const rotation = new Quaternion().setFromAxisAngle(
      getAxisVector(this.type).multiplyScalar(direction),
      delta,
    );
    quaternion.premultiply(rotation);

    this.target.rotation.setFromQuaternion(quaternion);
    if (this.onMove) {
      this.onMove(this.type, direction, delta);
    }
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    this.target.controlling = undefined;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
    if (this.onUp) {
      this.onUp();
    }
  };
}
