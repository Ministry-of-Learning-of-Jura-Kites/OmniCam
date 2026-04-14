import type { TresContext } from "@tresjs/core";
import type { IUserData } from "~/types/obj-3d-user-data";
import type { Group } from "three";
import { Raycaster, Vector2, Vector3 } from "three";
import type { ICamera } from "~/types/camera";
import type { SceneStates } from "~/types/scene-states";
import type { QuadrilateralPoints as QuadrilateralPoints } from "~/types/trapezoid";

type Axis = "x" | "y";

/**
 * Returns the local direction vector for a given axis based on our
 * "Normal = Z" basis convention.
 */
export function getGizmoLocalDirection(axis: Axis, direction: 1 | -1): Vector3 {
  // Convention:
  // Normal is Z (0,0,1)
  // X is (1,0,0)
  // Y is (0,1,0)
  return new Vector3(
    axis === "x" ? 1 : 0,
    axis === "y" ? 1 : 0,
    0,
  ).multiplyScalar(direction);
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

  private initialP: QuadrilateralPoints | undefined = undefined;

  private direction: 1 | -1;
  private parentGroup: Group;

  private isDragging = false;
  private startElev = new Vector3();
  private axisDir = new Vector3();
  private t0 = 0;
  constructor(
    axis: Axis,
    direction: 1 | -1,
    parentGroup: Group,
    faceId: string,
    cornerIndex: number,
    sceneStates: SceneStates,
    context: TresContext,
    yOffset: number,
  ) {
    this.axis = axis;
    this.direction = direction;
    this.parentGroup = parentGroup;
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

    // The starting point for the drag calculation
    this.startElev.set(p[0], p[1], p[2]);

    // --- THE FIX: Align with Gizmo Basis ---
    // Use the shared helper to get the vector (e.g., [1,0,0] for X)
    const localDir = getGizmoLocalDirection(this.axis, this.direction);

    // Transform local direction to World Space using the group's current orientation
    // This ensures that dragging moves exactly along the visible arrow's line
    this.axisDir = localDir
      .applyQuaternion(this.parentGroup.quaternion)
      .normalize();

    const ray = this.getRay(event);
    if (!ray) return;

    this.initialP = structuredClone(
      face.points.map(toRaw) as QuadrilateralPoints,
    );

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

    // Calculate the projection on the translation axis
    const t = this.closestTOnAxisLine(
      ray.origin,
      ray.direction,
      this.startElev,
      this.axisDir,
    );

    const delta = t - this.t0;

    // Compute the new position
    const newPos = this.startElev.clone().addScaledVector(this.axisDir, delta);

    this.sceneStates.facesManagement.updateCorner(
      this.faceId,
      this.cornerIndex,
      [newPos.x, newPos.y, newPos.z],
    );
  };

  private onPointerUp = () => {
    this.isDragging = false;

    const face = this.sceneStates.facesManagement.faces[this.faceId];

    if (face && this.initialP) {
      // If you want to ensure the plane stays perfectly flat (coplanar)
      // after a move, you would do math here.
      // But for a simple 4-point plane, we usually just accept the new position.

      const currentPoint = face.points[this.cornerIndex]!;

      this.sceneStates.facesManagement.updateCorner(
        this.faceId,
        this.cornerIndex,
        [currentPoint[0], currentPoint[1], currentPoint[2]],
      );
    }

    document.removeEventListener("pointermove", this.onPointerMove);
    document.removeEventListener("pointerup", this.onPointerUp);
  };
}
