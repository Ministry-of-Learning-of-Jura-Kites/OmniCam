// ~/composables/useAuth.ts
import { useState, useRuntimeConfig } from "#app";

export async function useAuth() {
  const config = useRuntimeConfig();
  const user = useState<string | null>("auth_user", () => null);

  try {
    const response = await $fetch<{ data: string | null }>(
      `http://${config.public.NUXT_PUBLIC_BACKEND_HOST}/api/v1/me`,
      {
        method: "GET",
        credentials: "include",
      },
    );
    user.value = response.data;
  } catch (err) {
    console.error("fetchModel error", err);
    user.value = null;
  }

  return { user };
}
