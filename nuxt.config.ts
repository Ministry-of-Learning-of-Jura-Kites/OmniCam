// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";
export default defineNuxtConfig({
  css: ["~/assets/css/tailwind.css"],
  vite: {
    plugins: [tailwindcss()],
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
      path: "~/components/ui",
      extensions: ["vue"],
    },
  ],
  routeRules: {
    "/projects": { redirect: "/" },
  },
  runtimeConfig: {
    public: {
      NUXT_PUBLIC_URL: process.env.NUXT_PUBLIC_URL,
    },
  },
});
