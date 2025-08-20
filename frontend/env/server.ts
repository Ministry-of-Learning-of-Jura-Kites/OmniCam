import { createEnv } from "@t3-oss/env-nuxt";
import { z } from "zod";

export const env = createEnv({
  server: {
    TEST: z.string().min(1),
  },
});
