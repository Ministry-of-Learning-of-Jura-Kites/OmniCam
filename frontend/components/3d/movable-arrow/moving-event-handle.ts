import type { IUserData } from "~/types/obj-3d-user-data";
import { MOVING_ARROW_CONFIG } from "~/constants";
import type { TresContext } from "@tresjs/core";
import { getAxisVector as createAxisVector } from "~/lib/utils";
import type { ICamera } from "~/types/camera";

export class MovingUserData implements IUserData {
  type: "x" | "y" | "z";
  // obj: ModelRef<THREE.Mesh>;
  cam: ICamera;
  context: TresContext;

  isDragging = false;

  constructor(type: string, cam: ICamera, context: TresContext) {
    this.type = type as "x" | "y" | "z";
    this.cam = cam;
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

    const camera = this.context.camera!;

    const point = this.cam.position.clone();

    const dirVector = createAxisVector(this.type);

    const vectorEnd = dirVector.add(point);

    const pointNDC = point.clone().project(camera.value!);
    const endNDC = vectorEnd.clone().project(camera.value!);

    const projectedVector = endNDC.sub(pointNDC).normalize();

    if (this.cam.position[this.type] != null) {
      const delta = createAxisVector(this.type).multiplyScalar(
        (projectedVector.x * event.movementX -
          projectedVector.y * event.movementY) *
          MOVING_ARROW_CONFIG.DRAGGING_SENTIVITY,
      );
      this.cam.position.add(delta);
    }
  };

  handlePointerUpEvent = (_event: PointerEvent) => {
    this.isDragging = false;
    document.removeEventListener("pointerup", this.handlePointerUpEvent);
    document.removeEventListener("pointermove", this.handlePointerMoveEvent);
  };
}
