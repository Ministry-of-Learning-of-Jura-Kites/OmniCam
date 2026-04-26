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

  async function fetchUser() {
    try {
      const resp = await getMe();
      user.value = resp.data;
      return resp.data;
    } catch (err) {
      console.error("fetch /me error", err);
      user.value = null;
      return null;
    }
  }

  async function postLogin(loginForm: LoginRequest) {
    const base = getApiBaseUrlWithProtocol("http", config);
    await $fetch<Response>(concatUrl("login", base).href, {
      method: "POST",
      body: loginForm,
      credentials: "include",
    });
    await fetchUser();
  }

  async function postRegister(registerForm: RegisterRequest) {
    const base = getApiBaseUrlWithProtocol("http", config);
    return await $fetch<Response>(concatUrl("register", base).href, {
      method: "POST",
      body: registerForm,
      credentials: "include",
    });
  }

  async function postLogout() {
    const base = getApiBaseUrlWithProtocol("http", config);

    await $fetch<null>(concatUrl(`logout`, base).href, {
      method: "POST",
      credentials: "include",
    });
    user.value = null;
  }

  async function getMe() {
    const headers = useRequestHeaders(["cookie"]);

    const base = getApiBaseUrlWithProtocol("http", config);
    return $fetch<{ data: User }>(concatUrl(`me`, base).href, {
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
