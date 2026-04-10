// ~/composables/useAuth.ts
import { useState, useRuntimeConfig } from "#app";
import { getApiBaseUrlWithProtocol } from "~/utils/url";

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  username: string;
}

export interface RegisterRequest {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  identifier: string;
  password: string;
}

export interface LoginRequest {
  identifier: string;
  password: string;
}

export interface User {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  password: string;
  createdAt: string;
  updatedAt: string;
}

export interface Response {
  data: User;
  token: string;
}

export function useAuth() {
  const config = useRuntimeConfig();

  const user = useState<User | null>("user", () => null);

  const { data, execute: fetchUser } = useAsyncData("auth-me", getMe, {
    immediate: false,
    transform: (resp) => resp.data,
    lazy: true,
  });

  watch(data, (data) => {
    user.value = data ?? null;
  });

  async function postLogin(loginForm: LoginRequest) {
    const base = getApiBaseUrlWithProtocol("http", config, false);
    await $fetch<Response>(new URL("login", base).href, {
      method: "POST",
      body: loginForm,
      credentials: "include",
    });
  }

  async function postRegister(registerForm: RegisterRequest) {
    const base = getApiBaseUrlWithProtocol("http", config, false);
    await $fetch<Response>(new URL("register", base).href, {
      method: "POST",
      body: registerForm,
      credentials: "include",
    });
  }

  async function postLogout() {
    const base = getApiBaseUrlWithProtocol("http", config, false);

    $fetch<null>(new URL(`logout`, base).href, {
      method: "POST",
      credentials: "include",
    });
  }

  async function getMe() {
    const headers = useRequestHeaders(["cookie"]);

    const base = getApiBaseUrlWithProtocol("http", config, true);
    return $fetch<{ data: User }>(new URL(`me`, base).href, {
      method: "GET",
      headers: headers,
      credentials: "include",
      onResponseError({ response }) {
        console.error("fetch /me error", response._data);
        user.value = null;
      },
    });
  }

  return { user, postLogin, postRegister, fetchUser, getMe, postLogout };
}
