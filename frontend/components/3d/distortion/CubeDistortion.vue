<!-- eslint-disable vue/valid-template-root -->
<script setup lang="ts">
import type { WebGLRenderer } from "three";
import { EffectComposer } from "three/addons/postprocessing/EffectComposer.js";
import { OutputPass } from "three/addons/postprocessing/OutputPass.js";
import { RenderPass } from "three/addons/postprocessing/RenderPass.js";
import { ShaderPass } from "three/examples/jsm/postprocessing/ShaderPass.js";
import type { WatchHandle } from "vue";
import { SCENE_STATES_KEY } from "~/constants/state-keys";

const { logger } = useLogger("CubeDistortion");

const sceneStates = inject(SCENE_STATES_KEY);
if (sceneStates == undefined) {
  throw new Error("Expect to be called within scene states provider");
}

const FisheyeShader = {
  uniforms: {
    tCube: { value: null },
    uIsFisheye: { value: false },
    uFov: { value: 60 },
    uAspectRatio: { value: 1.0 },
  },
  vertexShader: `varying vec2 vUv;
void main() {
  vUv = uv;
  gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
}`,
  fragmentShader: `uniform samplerCube tCube;
varying vec2 vUv;
uniform bool uIsFisheye;
uniform float uFov;
uniform float uAspectRatio;

void main() {
  // Map UV from [0, 1] to [-1, 1] (centered)
  vec2 p = vUv * 2.0 - 1.0;

  vec2 aspectP = vec2(p.x * uAspectRatio, p.y);

  // Calculate distance from center
  float r = length(aspectP);

  // Only render inside the "circle"
  if (uIsFisheye && r > 1.0) {
    discard; // Or return black: gl_FragColor = vec4(0,0,0,1);
  }

  float halfAngle = (uFov * 3.14159265) / 360.0;

  // Fisheye Mapping
  // 'r' is the angle from the forward vector.
  float theta = r * halfAngle;
  float phi = atan(aspectP.y, aspectP.x);

  // Convert to 3D direction vector
  vec3 dir = vec3(
    sin(theta) * cos(phi), // X
    sin(theta) * sin(phi), // Y
    -cos(theta)             // Z- (Forward)
  );

  // Sample the CubeMap
  gl_FragColor = textureCube(tCube, dir);
}`,
};

let stopWatch: WatchHandle | null = null;
let renderCallback: { off: () => void } | null = null;
let composer: EffectComposer | null = null;

onMounted(() => {
  stopWatch = watch(
    () =>
      [sceneStates.cubeCamera.value, sceneStates.tresContext.value] as const,
    ([cubeCam, tresContext]) => {
      if (renderCallback) {
        renderCallback.off();
        renderCallback = null;
      }

      if (cubeCam == undefined || tresContext == undefined) {
        return;
      }

      logger.info("Setting up cube distortion");

      const renderer = tresContext.renderer.instance as WebGLRenderer;

      composer = new EffectComposer(renderer);

      const renderPass = new RenderPass(
        tresContext.scene,
        tresContext.camera.activeCamera!,
      );
      composer.addPass(renderPass);

      renderer.setClearColor(0x000000, 1);

      const customPass = new ShaderPass(FisheyeShader, "tDiffuse");
      composer.addPass(customPass);

      composer.setPixelRatio(5 * window.devicePixelRatio);

      customPass.uniforms.uFov! = sceneStates.currentFov;
      customPass.uniforms.uIsFisheye! = sceneStates.currentIsFisheye;
      customPass.uniforms.uAspectRatio! = sceneStates.aspectRatio;
      const outputPass = new OutputPass();
      composer.addPass(outputPass);

      renderCallback = tresContext.renderer.onRender(() => {
        const isDistorting =
          sceneStates.currentDistEnabled.value &&
          sceneStates.transformingInfo.value == undefined;
        if (!isDistorting || !composer) {
          return;
        }
        if (cubeCam.parent != null) {
          cubeCam.update(renderer, tresContext.scene);
        } else {
          logger.error("Improperly disposed cube camera");
          if (renderCallback) {
            renderCallback.off();
          }
          return;
        }
        customPass.uniforms.tCube!.value = cubeCam.renderTarget.texture;

        composer.render();
      });

      if (stopWatch) {
        stopWatch();
      }
    },
    { immediate: true },
  );
});

onUnmounted(() => {
  if (stopWatch) stopWatch();

  if (renderCallback) renderCallback.off();

  if (composer != null) {
    composer.dispose();
  }
});
</script>
