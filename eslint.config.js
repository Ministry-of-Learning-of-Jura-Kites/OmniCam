// eslint.config.mjs
import withNuxt from "./.nuxt/eslint.config.mjs";

export default withNuxt([
  {
    files: ["frontend/components/ui/**/*.{ts,vue}"], // Shadcn files
    rules: {
      "vue/require-default-prop": "off",
    },
  },
]);
