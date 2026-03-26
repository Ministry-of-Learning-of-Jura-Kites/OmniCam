// import type { WebGLProgramParametersWithUniforms, WebGLRenderer } from "three";
// import { DistortionMode } from "~/messages/protobufs/backend_frontend_event";
// import type { SceneStatesWithHelper } from "~/types/scene-states";
// // import type { ICamera } from "~/types/camera";

// export function calcFisheyeStrength(
//   type: DistortionMode,
//   intensity: number,
//   fov: number,
// ): number {
//   const fovRad = (fov * Math.PI) / 180;
//   const scale = Math.tan(fovRad / 2);

//   switch (type) {
//     case DistortionMode.LINEAR:
//       return intensity * 0.5 * scale;

//     case DistortionMode.QUAD:
//       return intensity * 0.05 * (scale * scale);

//     case DistortionMode.ATAN:
//       return (intensity * 2.5) / scale;

//     case DistortionMode.EQUISOLID:
//       return intensity * (1.2 / scale);

//     case DistortionMode.ORTHO: {
//       const orthoFit = 3 / scale;
//       return 1.0 + intensity * (orthoFit - 1.0);
//     }

//     case DistortionMode.PERSPECTIVE: {
//       const f = 1.0 / Math.tan(fovRad / 2);
//       return (1.0 / f) * intensity;
//     }

//     case DistortionMode.NONE:
//     default:
//       return 0.0;
//   }
// }

// const updateMv = `
//     float r = length(mvPosition.xy);
//     switch (uDistortionMode) {
//         case 0: // None
//           break;
//         case ${DistortionMode.PERSPECTIVE}: {
//           if (r > 0.0 && uStrength!=0.0) {
//               float theta = atan(r * uStrength);
//               float dist = theta / (r * uStrength);
//               mvPosition.xy *= dist;
//           }
//           break;
//         }
//         case ${DistortionMode.ORTHO}: {
//           if (r > 0.0 && uStrength!=0.0) {
//               float theta = atan(r);
//               float dist = (sin(theta) * uStrength) / r;
//               mvPosition.xy *= dist;
//           }
//           break;
//         }
//         case ${DistortionMode.EQUISOLID}: {
//           if (r > 0.0 && uStrength!=0.0) {
//               float theta = atan(r);
//               float dist = (2.0 * sin(theta * 0.5) * uStrength) / r;
//               mvPosition.xy *= dist;
//           }
//           break;
//         }
//         case ${DistortionMode.ATAN}: {
//           if (r > 0.0 && uStrength!=0.0) {
//               float theta = atan(r * uStrength);
//               float dist = theta / r;
//               // Auto-zoom to maintain center composition
//               float zoom = 1.0 + (uStrength * 0.15);
//               mvPosition.xy *= dist * zoom;
//           }
//           break;
//         }
//         case ${DistortionMode.LINEAR}: {
//           float dist = 1.0 / (1.0 + r * uStrength);

//           mvPosition.xy *= dist;
//           mvPosition.z -= r * r * (uStrength * 0.5);
//           break;
//         }
//         case ${DistortionMode.QUAD}: {
//           // 1. Setup coordinates (-1.0 to 1.0)
//           vec2 pos = 2.0 * uv - 1.0;

//           // 2. Adjust for aspect ratio
//           float aspect = 1.0;
//           vec2 p = pos;
//           p.x *= aspect;
//           float l = length(p);

//           // 3. Apply Distortion Math
//           // Note: We apply the math to the vertex positions before projection
//           float r = l;

//           // Standard Brown-Conrady Barrel Distortion
//           float distortedR = (l * (1.0 + uStrength * l * l)) / (1.0 + uStrength);

//           // Mix based on strength
//           r = mix(l, distortedR, uStrength);

//           // 4. Calculate new position
//           // We reconstruct the point based on the new radius 'r'
//           float phi = atan(p.y, p.x);
//           vec2 newP;
//           newP.x = (r * cos(phi)) / aspect;
//           newP.y = r * sin(phi)*2.0;

//           // 5. Final Position Output
//           // Converting back from our local -1.0->1.0 space to clip space
//           // gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);

//           // To warp the actual geometry:
//           gl_Position = vec4(newP, 0.0, 1.0);
//         }
//     }
// `;

// // Experimental composable with 3d shader to distort
// export function useFisheye(sceneStates: SceneStatesWithHelper) {
//   function injectFisheye(
//     shader: WebGLProgramParametersWithUniforms,
//     _renderer: WebGLRenderer,
//   ) {
//     shader.uniforms!.uStrength = sceneStates.currentDistStrength;
//     shader.uniforms!.uDistortionMode = sceneStates.currentDistMode;
//     shader.vertexShader =
//       `uniform float uStrength;\n` +
//       `uniform int uDistortionMode;\n` +
//       shader.vertexShader;
//     shader.vertexShader = shader.vertexShader.replace(
//       "#include <project_vertex>",
//       `
//     #include <project_vertex>
//     ${updateMv}
//     gl_Position = projectionMatrix * mvPosition;
//     `,
//     );
//   }
//   return { injectFisheye };
// }
