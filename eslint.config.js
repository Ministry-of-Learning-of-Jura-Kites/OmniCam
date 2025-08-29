// eslint.config.mjs
import withNuxt from "./.nuxt/eslint.config.mjs";

export default withNuxt([
  {
    files: ["frontend/components/ui/**/*.{ts,vue}"], // Shadcn files
    rules: {
      "vue/require-default-prop": "off",
    },
  },
  {
    files: ["frontend/**/*.vue"],
    rules: {
      "vue/html-self-closing": [
        "error",
        {
          html: {
            void: "any",
            normal: "any",
            component: "always",
          },
          svg: "always",
          math: "always",
        },
      ],
    },
  },
]);
