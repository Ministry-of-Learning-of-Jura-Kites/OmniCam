import type { ShaderMaterialParameters } from "three";
import type { SceneStatesWithHelper } from "~/types/scene-states";
// import type { ICamera } from "~/types/camera";

export function getFisheyeStrength(
  type: "linear" | "quad" | "atan" | "equisolid" | "ortho" | "pers",
  intensity: number,
  fov: number,
): number {
  const fovRad = (fov * Math.PI) / 180;
  const scale = Math.tan(fovRad / 2);

  switch (type) {
    case "linear":
      // Linear radial compression
      return intensity * 0.5 * scale;

    case "quad":
      // Stereographic approximation (r^2)
      return intensity * 0.05 * (scale * scale);

    case "atan":
      // Equidistant mapping (r = f * theta)
      // High intensity needed to move into the curve of the atan function
      return (intensity * 2.5) / scale;

    case "equisolid":
      // r = 2f sin(theta/2). Preserves area.
      return intensity * (1.2 / scale);

    case "ortho":
      // r = f sin(theta). Extreme peep-hole.
      return intensity * (0.9 / scale);

    case "pers": {
      const f = 1.0 / Math.tan(fovRad / 2);
      return (1.0 / f) * intensity;
    }
    default:
      return 0.0;
  }
}

const updateMvQuad = `
    float dist = 1.0 / (1.0 + r * r * uStrength);

    mvPosition.xy *= dist;
    mvPosition.z -= r * r * uStrength;
`;

const updateMvLinear = `
    float dist = 1.0 / (1.0 + r * uStrength);

    mvPosition.xy *= dist;
    mvPosition.z -= r * r * (uStrength * 0.5);
`;

// Equidistant
const updateMvAtan = `
    if (r > 0.0 && uStrength!=0.0) {
        float theta = atan(r * uStrength); 
        float dist = theta / r;
        // Auto-zoom to maintain center composition
        float zoom = 1.0 + (uStrength * 0.15); 
        mvPosition.xy *= dist * zoom;
    }
`;

const updateMvEquisolid = `
    if (r > 0.0 && uStrength!=0.0) {
        float theta = atan(r); 
        float dist = (2.0 * sin(theta * 0.5) * uStrength) / r;
        mvPosition.xy *= dist;
    }
`;

const updateMvOrthogonal = `
    if (r > 0.0 && uStrength!=0.0) {
        float theta = atan(r);
        float dist = (sin(theta) * uStrength) / r;
        mvPosition.xy *= dist;
    }
`;

const updateMvPers = `
    if (r > 0.0 && uStrength!=0.0) {
        // 2. Convert perspective distance back to an angle (theta)
        // In perspective: r = f * tan(theta) -> theta = atan(r/f)
        // We treat uStrength as the inverse focal length (1/f)
        float theta = atan(r * uStrength);
        
        // 3. The Equidistant rule: r_distorted = f * theta
        // To find the multiplier: (f * theta) / r_original
        // Which simplifies to: theta / (uStrength * r)
        float dist = theta / (r * uStrength);
        
        // 4. Apply
        mvPosition.xy *= dist;
    }
`;

export function useFisheye(sceneStates: SceneStatesWithHelper) {
  function injectFisheye(shader: ShaderMaterialParameters) {
    shader.uniforms!.uStrength = sceneStates.fisheyeFovStrength;

    shader.vertexShader = `uniform float uStrength;\n` + shader.vertexShader;
    shader.vertexShader = shader.vertexShader.replace(
      "#include <project_vertex>",
      `
    #include <project_vertex>

    ${updateMvPers}

    gl_Position = projectionMatrix * mvPosition;
    `,
    );
  }
  return { injectFisheye };
}
