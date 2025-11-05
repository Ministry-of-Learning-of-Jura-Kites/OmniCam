import type { RuntimeConfig } from "nuxt/schema";

export function getHostFromRuntime(runtime: RuntimeConfig, isClient: boolean) {
  if (isClient) {
    return runtime.public.externalBackendHost;
  }
  return runtime.internalBackendHost;
}
