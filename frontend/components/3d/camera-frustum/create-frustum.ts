import {
  BufferGeometry,
  MathUtils,
  Vector3,
  Float32BufferAttribute,
} from "three";

export function createFrustumGeometry(
  fov: number,
  aspect: number,
  length = 10000,
): { mesh: BufferGeometry; lines: BufferGeometry } {
  const fovRad = MathUtils.degToRad(fov);
  const halfFovV = Math.tan(fovRad / 2);

  // length plane dimensions
  const farHeight = length * halfFovV * 2;
  const farWidth = farHeight * aspect;
  const farHalfW = farWidth / 2;
  const farHalfH = farHeight / 2;

  // Define the 8 corner vertices (in Camera Space: +Z is forward)
  // Note: We're setting up the geometry so the camera is at (0,0,0) looking down the +Z axis.
  const vertices: Vector3[] = [
    // Near Plane Corners (+Z)
    new Vector3(0, 0, 0),

    // length Plane Corners (+Z)
    new Vector3(farHalfW, farHalfH, -length), // 4: Top-Right
    new Vector3(-farHalfW, farHalfH, -length), // 5: Top-Left
    new Vector3(-farHalfW, -farHalfH, -length), // 6: Bottom-Left
    new Vector3(farHalfW, -farHalfH, -length), // 7: Bottom-Right
  ];

  // Line Highlight (Wireframe Edges)
  const lineGeometry = new BufferGeometry().setFromPoints(vertices);

  // Indices for LineSegments (4 edges connecting to apex, 4 edges on the far plane)
  const lineIndices = [
    // Far plane edges (4 segments)
    1, 2, 2, 3, 3, 4, 4, 1,
    // Edges connecting to the Apex (4 segments)
    0, 1, 0, 2, 0, 3, 0, 4,
  ];

  const positionAttribute = new Float32BufferAttribute(
    new Float32Array(vertices.flatMap((v) => v.toArray())),
    3,
  );
  lineGeometry.setAttribute("position", positionAttribute);
  lineGeometry.setIndex(lineIndices);

  // Color Filled Sides (Translucent Mesh)

  const meshGeometry = new BufferGeometry().setFromPoints(vertices);

  const meshIndices = [
    // Far face (1, 2, 3, 4) - Winding order reversed for visibility from the outside
    1,
    3,
    2, // Far Tri 1 (TR, BL, TL)
    1,
    4,
    3, // Far Tri 2 (TR, BR, BL)

    // Side faces (connecting Apex 0 to the far plane edges)
    // Top face (0, 1, 2)
    0,
    1,
    2,
    // Bottom face (0, 3, 4) - Note: Using 0, 4, 3 might be better for consistent winding
    0,
    3,
    4,
    // Left face (0, 2, 3)
    0,
    2,
    3,
    // Right face (0, 4, 1)
    0,
    4,
    1,
  ];

  meshGeometry.setAttribute("position", positionAttribute); // Reuse position attribute
  meshGeometry.setIndex(meshIndices);
  meshGeometry.computeVertexNormals(); // For proper lighting

  return { mesh: meshGeometry, lines: lineGeometry };
}
