import type { BufferGeometry } from "three";
import { CylinderGeometry } from "three";
import { mergeGeometries } from "three/addons/utils/BufferGeometryUtils.js";

export type CameraObjectGeo = "body" | "lens";

const store = {} as Record<CameraObjectGeo, BufferGeometry>;

export function useCamObjGeoCache() {
  function create(geoName: CameraObjectGeo) {
    switch (geoName) {
      case "body": {
        const body = new CylinderGeometry(0.2, 0.2, 0.5);
        body.rotateX(Math.PI / 2);
        body.translate(0, 0, 0.31);

        const backCyl = new CylinderGeometry(0.05, 0.05, 0.1);
        backCyl.rotateX(Math.PI / 2);
        backCyl.translate(0, 0, 0.6);

        const handleArm = new CylinderGeometry(0.05, 0.05, 0.25);
        handleArm.rotateX(Math.PI / 5);
        handleArm.translate(0, 0.07, 0.71);

        const handleTop = new CylinderGeometry(0.15, 0.15, 0.02);
        handleTop.rotateX(Math.PI / 5);
        handleTop.translate(0, 0.17, 0.78);
        console.log("called");
        return mergeGeometries([body, backCyl, handleArm, handleTop]);
      }
      case "lens": {
        const lens = new CylinderGeometry(0.05, 0.05, 0.12);
        lens.rotateX(Math.PI / 2);
        lens.translate(0, 0, 0.06);
        return lens;
      }
    }
  }

  function get(geoName: CameraObjectGeo) {
    if (store[geoName] == undefined) {
      const newGeo = create(geoName);
      store[geoName] = newGeo;
      return newGeo;
    }
    return store[geoName];
  }

  return { get };
}
