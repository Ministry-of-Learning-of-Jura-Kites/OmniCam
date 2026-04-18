import { onUnmounted, useId, markRaw } from "vue";
import type { BufferGeometry } from "three";

export interface GeoCacheConfig<T extends string> {
  name: string;
  create: (geoName: T) => BufferGeometry;
}

export function createGeoCache<T extends string>(config: GeoCacheConfig<T>) {
  const geoCache = {} as Record<string, [BufferGeometry, Set<string>]>;
  const idToGeo = {} as Record<string, Set<T>>;

  return function useCache() {
    const id = useId();

    function get(geoName: T): BufferGeometry {
      // 1. Register link between Component and Geo
      if (!idToGeo[id]) idToGeo[id] = new Set();
      idToGeo[id].add(geoName);

      // 2. Manage Cache
      if (!geoCache[geoName]) {
        console.log(`[${config.name}] Creating:`, geoName);
        const newGeo = markRaw(config.create(geoName));
        geoCache[geoName] = [newGeo, new Set([id])];
        return newGeo;
      }

      const [geo, members] = geoCache[geoName];
      members.add(id);
      return geo;
    }

    onUnmounted(() => {
      const registry = idToGeo[id];
      if (!registry) return;

      for (const geoKey of registry) {
        const entry = geoCache[geoKey];
        if (!entry) continue;

        const [geo, members] = entry;
        members.delete(id);

        if (members.size === 0) {
          console.log(`[${config.name}] Disposing:`, geoKey);
          geo.dispose();
          delete geoCache[geoKey];
        }
      }
      delete idToGeo[id];
    });

    return { get };
  };
}
