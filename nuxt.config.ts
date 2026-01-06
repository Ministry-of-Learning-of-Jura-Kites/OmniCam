// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";
import { templateCompilerOptions } from "@tresjs/core";

export default defineNuxtConfig({
  css: [
    "~/assets/css/tailwind.css",
    "@fortawesome/fontawesome-free/css/all.min.css",
  ],
  vite: {
    plugins: [
      {
        apply: "build",
        name: "vite-plugin-ignore-sourcemap-warnings",
        configResolved(config) {
          const originalOnWarn = config.build.rollupOptions.onwarn;
          config.build.rollupOptions.onwarn = (warning, warn) => {
            if (
              warning.code === "SOURCEMAP_BROKEN" &&
              warning.plugin === "@tailwindcss/vite:generate:build"
            ) {
              return;
            }

            if (originalOnWarn) {
              originalOnWarn(warning, warn);
            } else {
              warn(warning);
            }
          };
        },
      },
      tailwindcss(),
    ],
    vue: templateCompilerOptions,
  },
  compatibilityDate: "2025-07-15",
  ssr: true,
  devtools: { enabled: true },
  srcDir: "frontend",
  modules: ["@nuxt/eslint", "shadcn-nuxt"],
  app: {
    head: {
      link: [{ rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
    },
  },
  dir: {
    public: "frontend/public",
  },
  serverDir: "frontend/server",
  nitro: {
    plugins: ["plugins/runtime-env.server.ts"],
  },
  shadcn: {
    /**
     * Prefix for all the imported component
     */
    prefix: "",
    /**
     * Directory that the component lives in.
     * @default "./components/ui"
     */
    componentDir: "~/components/ui",
  },
  components: [
    {
      path: "~/components",
      extensions: ["vue"],
    },
  ],
  routeRules: {
    "/projects": { redirect: "/" },
  },
  runtimeConfig: {
    internalBackendHost: "error",
    public: {
      externalBackendHost: "error",
    },
  },
});
