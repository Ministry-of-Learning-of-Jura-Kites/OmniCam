import type { TresContext } from "@tresjs/core";
import type { IUserData } from "~/types/obj-3d-user-data";
import { Raycaster, Vector2, Vector3 } from "three";
import type { ICamera } from "~/types/camera";
import type { SceneStates } from "~/types/scene-states";

type Axis = "x" | "y" | "z";

function axisVector(axis: Axis) {
  switch (axis) {
    case "x":
      return new Vector3(1, 0, 0);
    case "y":
      return new Vector3(0, 1, 0);
    case "z":
      return new Vector3(0, 0, 1);
  }
}

export class CornerTranslateUserData implements IUserData {
  type = "corner-translate" as const;

  axis: Axis;
  faceId: string;
  cornerIndex: number;
  context: TresContext;
  sceneStates: SceneStates;
  yOffset: number;

  private raycaster = new Raycaster();
  private mouse = new Vector2();

  private isDragging = false;
  private startElev = new Vector3();
  private axisDir = new Vector3();
  private t0 = 0;

  constructor(
    axis: Axis,
    faceId: string,
    cornerIndex: number,
    sceneStates: SceneStates,
    context: TresContext,
    yOffset: number,
  ) {
    this.axis = axis;
    this.faceId = faceId;
    this.cornerIndex = cornerIndex;
    this.sceneStates = sceneStates;
    this.context = context;
    this.yOffset = yOffset;
  }

  target: unknown;
  cam!: ICamera;

  handleEvent(eventType: string, event: Event) {
    if (eventType === "pointerdown") {
      this.onPointerDown(event as PointerEvent);
    }
  }

  private getRay(event: PointerEvent) {
    const domElement = this.context.renderer?.instance?.domElement;
    const camera = this.context.camera?.activeCamera?.value;

    if (!domElement || !camera) return null;

    const rect = domElement.getBoundingClientRect();

    this.mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    this.mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

    this.raycaster.setFromCamera(this.mouse, camera);
    return this.raycaster.ray;
  }

  private closestTOnAxisLine(
    rayOrigin: Vector3,
    rayDir: Vector3,
    l0: Vector3,
    ld: Vector3,
  ) {
    const w0 = rayOrigin.clone().sub(l0);

    const a = rayDir.dot(rayDir);
    const b = rayDir.dot(ld);
    const c = ld.dot(ld);
    const d = rayDir.dot(w0);
    const e = ld.dot(w0);

    const den = a * c - b * b;
    if (Math.abs(den) < 1e-8) {
      return -e;
    }
    return (a * e - b * d) / den;
  }

  private onPointerDown(event: PointerEvent) {
    const face = this.sceneStates.facesManagement.faces[this.faceId];
    if (!face) return;

    const p = face.points?.[this.cornerIndex];
    if (!p) return;

    this.startElev.set(p[0], p[1] + this.yOffset, p[2]);
    this.axisDir = axisVector(this.axis).clone().normalize();

    const ray = this.getRay(event);
    if (!ray) return;

    this.t0 = this.closestTOnAxisLine(
      ray.origin,
      ray.direction,
      this.startElev,
      this.axisDir,
    );

    this.isDragging = true;
    document.addEventListener("pointermove", this.onPointerMove);
    document.addEventListener("pointerup", this.onPointerUp);
  }

  private onPointerMove = (event: PointerEvent) => {
    if (!this.isDragging) return;

    const ray = this.getRay(event);
    if (!ray) return;

    const t = this.closestTOnAxisLine(
      ray.origin,
      ray.direction,
      this.startElev,
      this.axisDir,
    );
    const delta = t - this.t0;

    const newElev = this.startElev.clone().addScaledVector(this.axisDir, delta);

    const newBase = newElev.clone();
    newBase.y -= this.yOffset;

    this.sceneStates.facesManagement.updateCorner(
      this.faceId,
      this.cornerIndex,
      [newBase.x, newBase.y, newBase.z],
    );
  };

  private onPointerUp = () => {
    this.isDragging = false;
    document.removeEventListener("pointermove", this.onPointerMove);
    document.removeEventListener("pointerup", this.onPointerUp);
  };
}
