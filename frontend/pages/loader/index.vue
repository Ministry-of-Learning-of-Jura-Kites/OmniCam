<script setup lang="ts">
import { ref, onMounted } from "vue";
import * as THREE from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader.js";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js"; // add this

const canvasRef = ref<HTMLCanvasElement | null>(null);

onMounted(() => {
  if (!canvasRef.value) return;

  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(
    20,
    canvasRef.value.clientWidth / canvasRef.value.clientHeight,
    1,
    400,
  );
  camera.position.z = 80;
  camera.position.x = 45;
  camera.position.y = -25;

  const renderer = new THREE.WebGLRenderer({
    canvas: canvasRef.value,
    antialias: true,
    // toneMapping: THREE.ACESFilmicToneMapping,
  });
  renderer.setSize(canvasRef.value.clientWidth, canvasRef.value.clientHeight);
  renderer.setClearColor(0xffffff, 1);
  renderer.outputColorSpace = THREE.SRGBColorSpace;
  renderer.toneMapping = THREE.ACESFilmicToneMapping;

  const ambientLight = new THREE.AmbientLight(0xffffff, 0.8);
  scene.add(ambientLight);

  const light = new THREE.DirectionalLight(0xffffff, 1);
  light.position.set(5, 5, 5);
  scene.add(light);

  const loader = new GLTFLoader();
  loader.load(
    "/model/poly.gltf",
    (gltf) => {
      scene.add(gltf.scene);

      // Check if materials have maps (textures)
      gltf.scene.traverse((child) => {
        if ((child as THREE.Mesh).isMesh) {
          const mesh = child as THREE.Mesh;
          const material = mesh.material as THREE.MeshStandardMaterial;

          if (material.map) {
            console.log("Texture loaded successfully:", material.map);
          } else {
            console.warn("No texture found on material:", material);
          }
        }
      });
    },
    (xhr) => {
      console.log(`Loading progress: ${(xhr.loaded / xhr.total) * 100}%`);
    },
    (error) => {
      console.error("Failed to load GLTF or its textures:", error);
    },
  );

  // --- CAMERA CONTROLS ---
  const controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true; // smooth camera
  controls.dampingFactor = 0.05;
  controls.target.set(0, 0, 0); // rotate around model center
  controls.update();

  // Optional: keyboard WASD to move camera
  document.addEventListener("keydown", (event) => {
    const step = 5;
    switch (event.code) {
      case "KeyW":
        camera.position.z -= step;
        break;
      case "KeyS":
        camera.position.z += step;
        break;
      case "KeyA":
        camera.position.x -= step;
        break;
      case "KeyD":
        camera.position.x += step;
        break;
      case "ArrowUp":
        camera.position.y += step;
        break;
      case "ArrowDown":
        camera.position.y -= step;
        break;
    }
  });

  function animate() {
    requestAnimationFrame(animate);
    controls.update(); // required for damping
    renderer.render(scene, camera);
  }
  animate();
});
</script>

<template>
  <canvas ref="canvasRef" class="w-full h-screen"/>
</template>
