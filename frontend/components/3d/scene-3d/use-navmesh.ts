// use-navmesh.ts
import { NavMesh, Polygon, Vector3 as YukaVec3 } from "yuka";
import { mergeGeometries } from "three/examples/jsm/utils/BufferGeometryUtils.js";
import type { GLTF } from "three-stdlib";
import type { BufferGeometry } from "three";
import { Mesh, Vector3 } from "three";

export function useNavmesh() {
  const navMesh = ref<NavMesh | null>(null);
  const isReady = ref(false);
  const walkablePolygonCount = ref(0);

  function buildFromGLTF(gltf: GLTF) {
    isReady.value = false;
    navMesh.value = null;
    walkablePolygonCount.value = 0;

    // ── Step 1: Collect geometry ─────────────────────────────────────
    const namedFloorGeos: BufferGeometry[] = [];
    const allGeos: BufferGeometry[] = [];

    gltf.scene.updateWorldMatrix(true, true); // ensure matrixWorld is baked

    gltf.scene.traverse((child) => {
      if (!(child instanceof Mesh)) return;

      const name = child.name.toLowerCase();
      const geo = child.geometry.clone();
      geo.applyMatrix4(child.matrixWorld);

      allGeos.push(geo);

      const isNamedFloor =
        name.includes("floor") ||
        name.includes("ground") ||
        name.includes("walkable") ||
        name.includes("pavement") ||
        name.includes("road") ||
        name.includes("terrain");

      if (isNamedFloor) {
        namedFloorGeos.push(geo);
      }
    });

    const sourceGeos = namedFloorGeos.length > 0 ? namedFloorGeos : allGeos;

    if (sourceGeos.length === 0) {
      console.warn("[NavMesh] No geometry found in GLTF scene.");
      return;
    }

    // ── Step 2: Merge all source geometry ───────────────────────────
    let merged: BufferGeometry;
    try {
      merged = mergeGeometries(sourceGeos, false);
    } catch (e) {
      console.error("[NavMesh] Failed to merge geometries:", e);
      return;
    }

    merged.computeVertexNormals();

    const posAttr = merged.attributes.position;
    if (!posAttr) {
      console.warn("[NavMesh] Merged geometry has no position attribute.");
      return;
    }

    // ── Step 3: Build Yuka Vector3 vertex list ───────────────────────
    const vertices: YukaVec3[] = [];
    for (let i = 0; i < posAttr.count; i++) {
      vertices.push(
        new YukaVec3(posAttr.getX(i), posAttr.getY(i), posAttr.getZ(i)),
      );
    }

    // ── Step 4: Build Yuka Polygon list (only upward-facing tris) ───
    const polygons: Polygon[] = [];

    const processTri = (iA: number, iB: number, iC: number) => {
      const v0 = vertices[iA]!;
      const v1 = vertices[iB]!;
      const v2 = vertices[iC]!;

      // Compute face normal in Yuka space
      const edge1 = new YukaVec3().subVectors(v1, v0);
      const edge2 = new YukaVec3().subVectors(v2, v0);
      const normal = new YukaVec3().crossVectors(edge1, edge2).normalize();

      // Skip walls and ceilings — only keep faces pointing mostly upward
      // Threshold 0.4 = ~66° from vertical, covers gentle ramps too
      if (normal.y < 0.4) return;

      // Skip degenerate (zero-area) triangles
      const area = new YukaVec3().crossVectors(edge1, edge2).length() * 0.5;
      if (area < 1e-6) return;

      const poly = new Polygon();
      // Yuka Polygon.fromContour expects an array of Vector3
      poly.fromContour([
        new YukaVec3(v0.x, v0.y, v0.z),
        new YukaVec3(v1.x, v1.y, v1.z),
        new YukaVec3(v2.x, v2.y, v2.z),
      ]);

      polygons.push(poly);
    };

    if (merged.index) {
      for (let i = 0; i < merged.index.count; i += 3) {
        processTri(
          merged.index.getX(i),
          merged.index.getX(i + 1),
          merged.index.getX(i + 2),
        );
      }
    } else {
      for (let i = 0; i < posAttr.count; i += 3) {
        processTri(i, i + 1, i + 2);
      }
    }

    if (polygons.length === 0) {
      console.warn(
        "[NavMesh] No walkable polygons after filtering. " +
          "If your model has no upward-facing faces, lower the normal.y threshold or check mesh names.",
      );
      return;
    }

    // ── Step 5: Build NavMesh ────────────────────────────────────────
    const nm = new NavMesh();
    nm.fromPolygons(polygons);

    navMesh.value = nm;
    isReady.value = true;
    walkablePolygonCount.value = polygons.length;

    console.log(`[NavMesh] Ready — ${polygons.length} walkable polygons.`);

    // Clean up merged geometry
    merged.dispose();
  }

  /**
   * Find a navmesh-safe path between two Three.js Vector3 points.
   *
   * Returns ordered waypoints that stay on the navmesh surface,
   * or null if the navmesh isn't ready or no path exists.
   */
  function findPath(from: Vector3, to: Vector3): Vector3[] | null {
    if (!navMesh.value || !isReady.value) return null;

    const yukaFrom = new YukaVec3(from.x, from.y, from.z);
    const yukaTo = new YukaVec3(to.x, to.y, to.z);

    // getClosestRegion snaps off-mesh points to the nearest polygon
    // so clicks slightly above/below the surface still work
    const fromRegion = navMesh.value.getClosestRegion(yukaFrom);
    const toRegion = navMesh.value.getClosestRegion(yukaTo);

    if (!fromRegion || !toRegion) return null;

    const path = navMesh.value.findPath(yukaFrom, yukaTo);

    if (!path || path.length === 0) return null;

    return path.map((p) => new Vector3(p.x, p.y, p.z));
  }

  /**
   * Snap a raw clicked point to the nearest point ON the navmesh surface.
   * Useful to correct raycasted points that land slightly off the mesh.
   */
  function snapToNavmesh(point: Vector3): Vector3 | null {
    if (!navMesh.value || !isReady.value) return null;

    const yukaPoint = new YukaVec3(point.x, point.y, point.z);
    const region = navMesh.value.getClosestRegion(yukaPoint);
    if (!region) return null;

    const clamped = new YukaVec3();
    navMesh.value.clampMovement(region, yukaPoint, yukaPoint, clamped);

    return new Vector3(clamped.x, clamped.y, clamped.z);
  }

  function dispose() {
    navMesh.value = null;
    isReady.value = false;
    walkablePolygonCount.value = 0;
  }

  return {
    navMesh,
    isReady,
    walkablePolygonCount,
    buildFromGLTF,
    findPath,
    snapToNavmesh,
    dispose,
  };
}
