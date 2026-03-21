<!-- eslint-disable vue/valid-template-root -->
<script setup lang="ts">
import type { WebGLRenderer } from "three";
import { EffectComposer } from "three/addons/postprocessing/EffectComposer.js";
import { OutputPass } from "three/addons/postprocessing/OutputPass.js";
import { RenderPass } from "three/addons/postprocessing/RenderPass.js";
import { ShaderPass } from "three/examples/jsm/postprocessing/ShaderPass.js";
import { SCENE_STATES_KEY } from "~/constants/state-keys";
import { DistortionMode } from "~/messages/protobufs/autosave_event";

const sceneStates = inject(SCENE_STATES_KEY);
if (sceneStates == undefined) {
  throw new Error("Expect to be called within scene states provider");
}

const FisheyeShader = {
  uniforms: {
    tDiffuse: { value: null }, // Intensity: 0.0 is flat, 1.0 is extreme
    uFov: { value: 150.0 }, // Camera FOV in degrees
    uAspectRatio: { value: 1.0 }, // Updated via: width / height
    uStrength: { value: 0.5 },
    uDistortionMode: { value: DistortionMode.NONE },
    uIsFisheye: { value: true },
  },
  vertexShader: `
    varying vec2 vUv;
    void main() {
      vUv = uv;
      gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
    }
  `,
  fragmentShader: `
precision highp float;

uniform sampler2D tDiffuse;
uniform float uFov;         
uniform float uStrength;    
uniform bool uIsFisheye;    
uniform int uDistortionMode; 
uniform float uAspectRatio; 

varying vec2 vUv;

const float PI = 3.1415926535;

void main() {
    // 1. Setup coordinates (-1.0 to 1.0)
    vec2 pos = 2.0 * vUv - 1.0;
    
    // Adjust for aspect ratio so the radial math stays circular
    float aspect = uAspectRatio;
    vec2 p = pos;
    p.x *= aspect;
    float l = length(p);

    // 2. Circle Mask logic (Only if uIsFisheye is explicitly true)
    if(uIsFisheye && l > 1.0) {
        gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
        return;
    }

    // 3. Mathematical Constants
    float halfAperture = uFov * (PI / 360.0);
    float r = l; 

    float strengthWeight=1.0;

    // 4. Projection Logic via Template Strings
    // These calculate the 'r' (the distance to sample from in the source texture)
    switch(uDistortionMode) {
      
      case (${DistortionMode.PERSPECTIVE}): {
          // Pure Rectilinear (No distortion)
          r = l;
          break;
      }

      case (${DistortionMode.ORTHO}): {
          // Orthographic: Standard 'sphere' mapping
          r = asin(l * sin(halfAperture)) / halfAperture;
          break;
      }

      case (${DistortionMode.EQUISOLID}): {
          // Equisolid: Modern lens mapping (Area-preserving)
          r = (2.0 * asin(l * sin(halfAperture * 0.5))) / halfAperture;
          strengthWeight=3.0;
          break;
      }

      case (${DistortionMode.ATAN}): {
          // The "Bulge" - heavy center distortion
          r = atan(l * uStrength) / max(atan(uStrength), 0.0001);
          break;
      }

      case (${DistortionMode.LINEAR}): {
          // Equidistant: r = f * theta
          r = l; 
          break;
      }

      case (${DistortionMode.QUAD}): {
          // Standard Brown-Conrady Barrel Distortion
          // This affects the whole frame without a circle mask
          r = (l * (1.0 + uStrength * l * l)) / (1.0 + uStrength);
          strengthWeight=0.5;
          break;
      }
    }

    // Apply global strength blend
    r = mix(l, r, uStrength*strengthWeight);

    // 5. Remap to UV Space
    float phi = atan(p.y, p.x);
    vec2 uv;
    // We divide back by aspect ratio to return to 0.0 - 1.0 range
    uv.x = (r * cos(phi) / aspect) * 0.5 + 0.5;
    uv.y = (r * sin(phi)) * 0.5 + 0.5;

    // 6. Sampling & Bounds Check
    if (uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0) {
        gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
    } else {
        // Subtle edge softening only for the fisheye circle
        float edgeAlpha = uIsFisheye ? smoothstep(1.0, 0.98, l) : 1.0;
        vec4 color = texture2D(tDiffuse, uv);
        gl_FragColor = vec4(color.rgb * edgeAlpha, 1.0);
    }
}`,
};

onMounted(() => {
  watch(
    sceneStates.tresContext,
    (tresContext) => {
      if (tresContext == undefined) {
        console.log("This shouldn't happen");
        return;
      }

      const renderer = tresContext.renderer.instance as WebGLRenderer;

      const composer = new EffectComposer(renderer);

      const renderPass = new RenderPass(
        tresContext.scene,
        tresContext.camera.activeCamera!,
      );
      composer.addPass(renderPass);

      renderer.setClearColor(0x000000, 1);

      const customPass = new ShaderPass(FisheyeShader, "tDiffuse");
      composer.addPass(customPass);

      composer.setPixelRatio(5 * window.devicePixelRatio);

      customPass.uniforms.uStrength! = sceneStates.currentDistStrength;
      customPass.uniforms.uFov! = sceneStates.currentFov;
      customPass.uniforms.uAspectRatio! = sceneStates.aspectRatio;
      customPass.uniforms.uIsFisheye! = sceneStates.currentIsFisheye;
      customPass.uniforms.uDistortionMode! = sceneStates.currentDistMode;
      const outputPass = new OutputPass();
      composer.addPass(outputPass);

      tresContext.renderer.onRender(() => {
        if (sceneStates.currentDistMode.value !== DistortionMode.NONE) {
          composer.render();
        }
      });
    },
    { immediate: true, once: true },
  );
});
</script>
