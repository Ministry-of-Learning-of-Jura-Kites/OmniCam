// ~/composables/useAuth.ts
import { useState, useRuntimeConfig } from "#app";

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  username: string;
}

export async function useAuth() {
  const config = useRuntimeConfig();
  const user = useState<User | null>("auth_user", () => null);

  try {
    const response = await $fetch<{ data: User }>(
      `http://${config.public.backendHost}/api/v1/me`,
      {
        method: "GET",
        credentials: "include",
      },
    );
    user.value = response.data;
  } catch (err) {
    console.error("fetch /me error", err);
    user.value = null;
  }

  return { user };
}
