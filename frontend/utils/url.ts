import type { RuntimeConfig } from "nuxt/schema";

export function getApiProtocol(
  type: "http" | "websocket",
  config: RuntimeConfig,
) {
  switch (type) {
    case "http":
      return config.public.nuxtBackendSecure ? "http" : "https";
    case "websocket":
      return config.public.nuxtBackendSecure ? "ws" : "wss";
  }
}

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
  const protocol = getApiProtocol(type, config);
  const host = getHostFromRuntime(config, import.meta.client);
  const apiBase = addTrailingSlash(host);
  return new URL(`${protocol}://${apiBase}`);
}
