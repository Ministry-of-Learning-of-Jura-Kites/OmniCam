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

export function getApiBaseUrlWithProtocol(
  type: "http" | "websocket",
  config: RuntimeConfig,
  allowServerSide: boolean = false,
) {
  const protocol = getApiProtocol(type, config);
  let host = config.public.externalBackendHost;
  if (allowServerSide) {
    host = getHostFromRuntime(config, import.meta.client);
  }
  const apiBase = addTrailingSlash(host);
  return new URL(`${protocol}://${apiBase}`);
}
