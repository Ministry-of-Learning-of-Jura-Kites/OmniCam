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

  await useAsyncData(
    "auth-me",
    () =>
      $fetch<{ data: User }>(
        `http://${getHostFromRuntime(config, import.meta.client)}/api/v1/me`,
        {
          method: "GET",
          headers: headers,
          credentials: "include",
          onResponseError({ response }) {
            console.error("fetch /me error", response._data);
            user.value = null;
          },
        },
      ),
    {
      // Only run if we don't already have a user in state
      immediate: !user.value,
      transform: (res) => {
        user.value = res.data;
        return res.data;
      },
    },
  );

  return { user };
}
