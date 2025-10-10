import * as THREE from "three";

export function createFrustumGeometry(
  fov: number,
  aspect: number,
  length = 10000,
): THREE.BufferGeometry {
  const angle = THREE.MathUtils.degToRad(fov / 2);
  const height = Math.tan(angle) * length;
  const width = height * aspect;

  const vertices = new Float32Array([
    // V0: Apex
    0,
    0,
    0,

    // V1: Top-Left Corner of Base
    -width,
    height,
    -length,
    // V2: Top-Right Corner of Base
    width,
    height,
    -length,
    // V3: Bottom-Right Corner of Base
    width,
    -height,
    -length,
    // V4: Bottom-Left Corner of Base
    -width,
    -height,
    -length,
  ]);

  const indices = new Uint16Array([
    // Side 1: V0, V1, V2
    0, 1, 2,
    // Side 2: V0, V2, V3
    0, 2, 3,
    // Side 3: V0, V3, V4
    0, 3, 4,
    // Side 4: V0, V4, V1
    0, 4, 1,

    // Base
    1, 4, 3, 1, 3, 2,
  ]);

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.BufferAttribute(vertices, 3));
  geometry.setIndex(new THREE.BufferAttribute(indices, 1));
  geometry.computeVertexNormals();

  return geometry;
}
