import { createEnv } from "@t3-oss/env-nuxt";

// Server-side-only env
export const env = createEnv({
  server: {
    // TEST: z.string().min(1),
  },
});

console.info("Server env validated");
