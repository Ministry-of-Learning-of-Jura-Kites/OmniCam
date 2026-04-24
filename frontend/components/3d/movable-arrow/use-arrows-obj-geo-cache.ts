import { CylinderGeometry } from "three/webgpu";
import { mergeGeometries } from "three/addons/utils/BufferGeometryUtils.js";
import { MOVING_ARROW_CONFIG } from "~/constants";

// useCamObjGeoCache.ts
export type ArrowObjectGeo = "arrow";

export const useArrowObjGeoCache = createGeoCache<ArrowObjectGeo>({
  name: "MovableArrow",
  create(geoName) {
    switch (geoName) {
      case "arrow": {
        const { HEAD_RADIUS, HEAD_LENGTH, CYLINDER_RADIUS, CYLINDER_LENGTH } =
          MOVING_ARROW_CONFIG;

        // 1. The Central Shaft
        const cylinder = new CylinderGeometry(
          CYLINDER_RADIUS,
          CYLINDER_RADIUS,
          CYLINDER_LENGTH,
          8,
        );

        // 2. The Top Head (Pointing Up)
        const upHead = new CylinderGeometry(0, HEAD_RADIUS, HEAD_LENGTH, 8);
        const upOffset = (CYLINDER_LENGTH + HEAD_LENGTH) / 2;
        upHead.translate(0, upOffset, 0);

        // 3. The Bottom Head (Pointing Down)
        const downHead = new CylinderGeometry(HEAD_RADIUS, 0, HEAD_LENGTH, 8);
        const downOffset = -(CYLINDER_LENGTH + HEAD_LENGTH) / 2;
        downHead.translate(0, downOffset, 0);

        // Merge into a single BufferGeometry
        return mergeGeometries([cylinder, upHead, downHead]);
      }
    }
  },
});
