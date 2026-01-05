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

  const headers = useRequestHeaders(["cookie"]);

  const { data, error } = await useAsyncData(
    "auth-me",
    () => {
      return $fetch<{ data: User }>(
        `http://${getHostFromRuntime(config, import.meta.client)}/api/v1/me`,
        {
          method: "GET",
          headers: headers,
          credentials: "include",
        },
      );
    },
    {
      // Only run if we don't already have a user in state
      immediate: !user.value,
    },
  );
  if (data.value) {
    user.value = data.value.data;
  }

  if (error.value) {
    console.error("fetch /me error", error.value);
    user.value = null;
  }

  return { user };
}
