import type { RuntimeConfig } from "nuxt/schema";
import { getUrlFromRuntime } from "./get-host";

export function addTrailingSlash(host: string | URL) {
  const urlString = host instanceof URL ? host.href : host;

  return urlString.endsWith("/") ? urlString : `${urlString}/`;
}

export function concatUrl(extended: string | URL, base: string | URL) {
  return new URL(extended, addTrailingSlash(base));
}

export function getApiBaseUrlWithProtocol(
  type: "http" | "websocket",
  config: RuntimeConfig,
) {
  let baseUrl: string;
  if (type == "http") {
    baseUrl = getUrlFromRuntime(config, import.meta.client);
  } else {
    baseUrl = config.public.externalWsBackendUrl;
  }
  const apiBase = addTrailingSlash(baseUrl);
  return new URL(apiBase);
}
