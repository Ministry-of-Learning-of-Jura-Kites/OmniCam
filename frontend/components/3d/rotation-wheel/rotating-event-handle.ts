import type { TresContext } from "@tresjs/core";
import type { ICamera } from "~/types/camera";
import type { IUserData } from "~/types/obj-3d-user-data";
import * as THREE from "three";
import { getAxisVector } from "~/lib/utils";

export const ROTATING_TYPE = "rotation";

export class RotatingUserData implements IUserData {
  type: "x" | "y" | "z";
  cam: ICamera;
  context: TresContext;

  isDragging = false;

  initialCamQuaternion = new THREE.Quaternion();
  downAngle = 0;

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

    this.initialCamQuaternion.setFromEuler(this.cam.rotation);

    const ele = this.context.renderer.value.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height) * 2 + 1;
    const objNdc = this.cam.position
      .clone()
      .project(this.context.camera.value!);
    this.downAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);
  };

  handlePointerMoveEvent = (event: PointerEvent) => {
    const objNdc = this.cam.position.clone();
    objNdc.project(this.context.camera.value!);

    const ele = this.context.renderer.value.domElement;
    const rect = ele.getBoundingClientRect();

    const scaledX = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    const scaledY = -((event.clientY - rect.top) / rect.height) * 2 + 1;

    const moveAngle = Math.atan2(scaledY - objNdc.y, scaledX - objNdc.x);

    const delta = moveAngle - this.downAngle!;

    const cameraDirection = new THREE.Vector3();

    this.context.camera.value!.getWorldDirection(cameraDirection);

    // Clockwise/Anticlockwise depend on side of object looked from
    const direction = -Math.sign(cameraDirection[this.type]);

    const quaternion = this.initialCamQuaternion.clone();
    const rotation = new THREE.Quaternion().setFromAxisAngle(
      getAxisVector(this.type).multiplyScalar(direction),
      delta,
    );
    quaternion.premultiply(rotation);

    this.cam.rotation.setFromQuaternion(quaternion);
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    this.cam.controlling = undefined;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
