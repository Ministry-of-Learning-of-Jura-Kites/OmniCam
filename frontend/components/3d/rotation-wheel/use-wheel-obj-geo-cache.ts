import { TorusGeometry } from "three/webgpu";
import { ROTATING_TORUS_CONFIG } from "~/constants";

export type WheelObjectGeo = "wheel";

export const useWheelObjGeoCache = createGeoCache<WheelObjectGeo>({
  name: "RotationWheel",
  create(geoName) {
    switch (geoName) {
      case "wheel": {
        const geometry = new TorusGeometry(
          ROTATING_TORUS_CONFIG.RADIUS,
          ROTATING_TORUS_CONFIG.TUBE_RADIUS,
          16,
          100,
        );

        // Merge into a single BufferGeometry
        return geometry;
      }
    }
  },
});
