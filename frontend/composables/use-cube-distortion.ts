import {
  CubeCamera,
  type Euler,
  Matrix3,
  type Scene,
  type Vector3,
  WebGLCubeRenderTarget,
  type WebGLRenderer,
  type Camera,
} from "three";

import { EffectComposer } from "three/examples/jsm/postprocessing/EffectComposer.js";
import { OutputPass } from "three/examples/jsm/postprocessing/OutputPass.js";
import { ShaderPass } from "three/examples/jsm/postprocessing/ShaderPass.js";

export const FisheyeCubeShader = {
  uniforms: {
    tCube: { value: null },
    uFov: { value: 60 },
    uAspectRatio: { value: 1.0 },
    uCameraRotation: { value: new Matrix3() },
    uIsFisheye: { value: true },
  },

  vertexShader: `
    varying vec2 vUv;

    void main() {
      vUv = uv;

      gl_Position =
        projectionMatrix *
        modelViewMatrix *
        vec4(position, 1.0);
    }
  `,

  fragmentShader: `
    uniform samplerCube tCube;
    uniform float uFov;
    uniform float uAspectRatio;
    uniform mat3 uCameraRotation;
    uniform bool uIsFisheye;
    varying vec2 vUv;
    const float PI = 3.14159265359;

    void main() {
      vec2 p = vUv * 2.0 - 1.0;

      // same projection always
      vec2 pAspect = vec2(p.x * uAspectRatio, p.y);
      float r = length(pAspect);

      float halfAngle = radians(uFov * 0.5);
      vec2 dir2D = (r > 0.0001) ? normalize(pAspect) : vec2(0.0, 0.0);
      float theta = r * halfAngle;

      vec3 dir = vec3(
        sin(theta) * dir2D.x,
        sin(theta) * dir2D.y,
        -cos(theta)
      );

      vec3 worldDir = uCameraRotation * dir;
      gl_FragColor = textureCube(tCube, worldDir);

      // fisheye: just black out outside the circle in screen center
      if (uIsFisheye && length(vec2(p.x * uAspectRatio, p.y)) > 1.0) {
        gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
      }
    }
  `,
};

export function createCubeDistortionRenderer(
  renderer: WebGLRenderer,
  scene: Scene,
) {
  const cubeRenderTarget = new WebGLCubeRenderTarget(1024);
  const cubeCamera = new CubeCamera(0.1, 1000, cubeRenderTarget);
  const composer = new EffectComposer(renderer);
  const fisheyePass = new ShaderPass(FisheyeCubeShader);

  composer.addPass(fisheyePass);
  composer.addPass(new OutputPass());

  function render({
    position,
    // rotation,
    fov,
    aspectRatio,
    width,
    height,
    isFisheye,
    activeCamera,
  }: {
    position: Vector3;
    rotation: Euler;
    fov: number;
    aspectRatio: number;
    width: number;
    height: number;
    isFisheye: boolean;
    activeCamera: Camera; // Added type
  }) {
    const oldAutoClear = renderer.autoClear;
    const oldRenderTarget = renderer.getRenderTarget();
    const oldXrEnabled = renderer.xr.enabled;

    renderer.xr.enabled = false;
    renderer.autoClear = true;

    // Update cube map
    cubeCamera.position.copy(position);
    // cubeCamera.rotation.copy(rotation);
    cubeCamera.update(renderer, scene);

    // Update Shader Uniforms
    fisheyePass.uniforms.tCube!.value = cubeRenderTarget.texture;
    fisheyePass.uniforms.uFov!.value = fov;
    fisheyePass.uniforms.uAspectRatio!.value = aspectRatio;
    fisheyePass.uniforms.uIsFisheye!.value = isFisheye;

    activeCamera.updateMatrixWorld();
    fisheyePass.uniforms
      .uCameraRotation!.value.setFromMatrix4(activeCamera.matrixWorld)
      .invert();

    composer.setSize(width, height);
    composer.render();

    // Restore state
    renderer.setRenderTarget(oldRenderTarget);
    renderer.xr.enabled = oldXrEnabled;
    renderer.autoClear = oldAutoClear;
  }

  function dispose() {
    composer.dispose();
    cubeRenderTarget.dispose();
  }

  return { render, dispose, cubeCamera, composer, fisheyePass };
}
