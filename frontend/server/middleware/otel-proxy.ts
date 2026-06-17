export default defineEventHandler((event) => {
  const config = useRuntimeConfig();

  const path = event.path;

  if (!path.startsWith(config.public.otelProxyPath)) {
    return;
  }

  const targetPath = path.slice(config.public.otelProxyPath.length) || "/";
  return proxyRequest(
    event,
    `${config.internalOtelProxyEndpoint}${targetPath}`,
  );
});
