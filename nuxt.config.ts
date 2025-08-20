// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  ssr: true,
  devtools: { enabled: true },
  srcDir: "frontend",
  modules: [
    "@nuxtjs/tailwindcss",
    "@nuxt/eslint",
    ...(import.meta.dev ? ["@scalar/nuxt"] : []),
  ],
  ...(import.meta.dev && {
    scalar: {
      pathRouting: {
        basePath: "/docs",
      },
      url: "/openapi.yaml",
    },
  }),
  app: {
    head: {
      link: [{ rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
    },
  },
  dir: {
    public: "frontend/public",
  },
  serverDir: "frontend/server",
});
