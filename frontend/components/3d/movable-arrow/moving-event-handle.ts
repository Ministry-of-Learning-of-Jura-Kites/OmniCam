import type { IUserData } from "~/types/obj-3d-user-data";
import { MOVING_ARROW_CONFIG } from "~/constants";
import type { TresContext } from "@tresjs/core";
import { getAxisVector as createAxisVector } from "~/lib/three";
import type { MovableObject } from "~/types/movable";

export const MOVING_TYPE = "moving";

export class MovingUserData implements IUserData<MovableObject> {
  type: "x" | "y" | "z";
  target: MovableObject;
  context: TresContext;

  onchange: undefined | (() => void);

  isDragging = false;

  constructor(
    type: string,
    target: MovableObject,
    context: TresContext,
    onchange?: () => void,
  ) {
    this.type = type as "x" | "y" | "z";
    this.target = target;
    this.context = context;
    this.onchange = onchange;
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
    this.target.controlling = {
      type: MOVING_TYPE,
      direction: this.type,
    };
    document.addEventListener("pointermove", this.handlePointerMoveEvent);
    document.addEventListener("pointerup", this.handlePointerUpEvent);
  };

  handlePointerMoveEvent = (event: PointerEvent) => {
    if (!this.isDragging) {
      return;
    }

    const camera = this.context.camera!;

    const point = this.target.position.clone();

    const dirVector = createAxisVector(this.type);

    const vectorEnd = dirVector.add(point);

    const pointNDC = point.clone().project(camera.value!);
    const endNDC = vectorEnd.clone().project(camera.value!);

    const projectedVector = endNDC.sub(pointNDC).normalize();

    if (this.target.position[this.type] != null) {
      const delta = createAxisVector(this.type).multiplyScalar(
        (projectedVector.x * event.movementX -
          projectedVector.y * event.movementY) *
          MOVING_ARROW_CONFIG.DRAGGING_SENTIVITY,
      );
      this.target.position.add(delta);
      if (this.onchange) {
        this.onchange();
      }
    }
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    this.target.controlling = undefined;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
