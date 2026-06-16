import {
  useCamObjGeoCache,
  type CameraObjectGeo,
} from "~/components/3d/camera-object-instance/use-cam-obj-geo-cache";

export type ModuleName = "CAMERA_OBJECT";

export type UniversalGeoName = CameraObjectGeo;

// TODO: Prevent mem leak
export function useGeoCache() {
  const camObj = useCamObjGeoCache();
  function get(module: ModuleName, geoName: UniversalGeoName) {
    switch (module) {
      case "CAMERA_OBJECT":
        return camObj.get(geoName as CameraObjectGeo);
    }
  }

  return { get };
}
