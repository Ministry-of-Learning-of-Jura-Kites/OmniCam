import type { WebGLProgramParametersWithUniforms, WebGLRenderer } from "three";
import { DistortionMode } from "~/messages/protobufs/autosave_event";
import type { SceneStatesWithHelper } from "~/types/scene-states";
// import type { ICamera } from "~/types/camera";

export function calcFisheyeStrength(
  type: DistortionMode,
  intensity: number,
  fov: number,
): number {
  const fovRad = (fov * Math.PI) / 180;
  const scale = Math.tan(fovRad / 2);

  switch (type) {
    case DistortionMode.LINEAR:
      return intensity * 0.5 * scale;

    case DistortionMode.QUAD:
      return intensity * 0.05 * (scale * scale);

    case DistortionMode.ATAN:
      return (intensity * 2.5) / scale;

    case DistortionMode.EQUISOLID:
      return intensity * (1.2 / scale);

    case DistortionMode.ORTHO:
      return intensity * (0.9 / scale);

    case DistortionMode.PERSPECTIVE: {
      const f = 1.0 / Math.tan(fovRad / 2);
      return (1.0 / f) * intensity;
    }

    case DistortionMode.NONE:
    default:
      return 0.0;
  }
}

const updateMv = `
    float r = length(mvPosition.xy);
    switch (uDistortionMode) {
        case 0: // None
          break;
        case ${DistortionMode.PERSPECTIVE}: {
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
          break;
        }
        case ${DistortionMode.ORTHO}: {
          if (r > 0.0 && uStrength!=0.0) {
              float theta = atan(r);
              float dist = (sin(theta) * uStrength) / r;
              mvPosition.xy *= dist;
          }
          break;
        }
        case ${DistortionMode.EQUISOLID}: {
          if (r > 0.0 && uStrength!=0.0) {
              float theta = atan(r); 
              float dist = (2.0 * sin(theta * 0.5) * uStrength) / r;
              mvPosition.xy *= dist;
          }
          break;
        }
        case ${DistortionMode.ATAN}: {
          if (r > 0.0 && uStrength!=0.0) {
              float theta = atan(r * uStrength); 
              float dist = theta / r;
              // Auto-zoom to maintain center composition
              float zoom = 1.0 + (uStrength * 0.15); 
              mvPosition.xy *= dist * zoom;
          }
          break;
        }
        case ${DistortionMode.LINEAR}: {
          float dist = 1.0 / (1.0 + r * uStrength);

          mvPosition.xy *= dist;
          mvPosition.z -= r * r * (uStrength * 0.5);
          break;
        }
        case ${DistortionMode.QUAD}: {
          float dist = 1.0 / (1.0 + r * r * uStrength);
          mvPosition.xy *= dist;
          mvPosition.z -= r * r * uStrength;
          break;
        }
    }
`;

export function useFisheye(sceneStates: SceneStatesWithHelper) {
  function injectFisheye(
    shader: WebGLProgramParametersWithUniforms,
    _renderer: WebGLRenderer,
  ) {
    shader.uniforms!.uStrength = sceneStates.distortionStrength;
    shader.uniforms!.uDistortionMode = sceneStates.distortionMode;

    shader.vertexShader =
      `uniform float uStrength;\n` +
      `uniform int uDistortionMode;\n` +
      shader.vertexShader;
    shader.vertexShader = shader.vertexShader.replace(
      "#include <project_vertex>",
      `
    #include <project_vertex>

    ${updateMv}

    gl_Position = projectionMatrix * mvPosition;
    `,
    );
  }
  return { injectFisheye };
}
