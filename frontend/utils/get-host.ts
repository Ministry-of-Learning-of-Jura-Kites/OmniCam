import type { RuntimeConfig } from "nuxt/schema";

export function getUrlFromRuntime(runtime: RuntimeConfig, isClient: boolean) {
  if (isClient) {
    return runtime.public.externalBackendUrl;
  }
  return runtime.internalBackendUrl;
}
