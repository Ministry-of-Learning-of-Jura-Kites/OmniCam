<script setup lang="ts">
definePageMeta({
  layout: false,
});
interface RegisterRequest {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  password: string;
}

interface LoginRequest {
  identifier: string;
  password: string;
}

interface User {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  password: string;
  createdAt: string;
  updatedAt: string;
}

interface Response {
  data: User;
  token: string;
}

const config = useRuntimeConfig();

const activeTab = ref<"signup" | "signin">("signup");

const registerForm = reactive<RegisterRequest>({
  firstName: "",
  lastName: "",
  username: "",
  email: "",
  password: "",
});

const touched = reactive({
  firstName: false,
  lastName: false,
  username: false,
  email: false,
  password: false,
});

const loginForm = reactive<LoginRequest>({
  identifier: "",
  password: "",
});

async function register() {
  // console.log(registerForm);
  try {
    const _response = await $fetch<Response>(
      "http://" + config.public.NUXT_PUBLIC_BACKEND_HOST + "/api/v1/register",
      {
        method: "POST",
        body: registerForm,
        credentials: "include",
      },
    );
    navigateTo("/");
  } catch (err) {
    console.log(err);
  }

  return;
}

async function login() {
  // console.log(loginForm);
  try {
    const _response = await $fetch<Response>(
      "http://" + config.public.NUXT_PUBLIC_BACKEND_HOST + "/api/v1/login",
      {
        method: "POST",
        body: loginForm,
        credentials: "include",
      },
    );
    navigateTo("/");
  } catch (err) {
    console.log(err);
  }

  return;
}

function markTouched(field: keyof RegisterRequest) {
  touched[field] = true;
}

function checkPasswordFormat(password: string): boolean {
  if (password.length < 8 || password.length > 255) return false;

  let hasNumber = false;
  let hasSymbol = false;

  for (const letter of password) {
    if (/\d/.test(letter)) hasNumber = true;
    if (/[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?~`]/.test(letter)) hasSymbol = true;
  }

  return hasNumber && hasSymbol;
}

const errors = computed(() => {
  return {
    firstName:
      touched.firstName && !registerForm.firstName
        ? "First name is required"
        : "",
    lastName:
      touched.lastName && !registerForm.lastName ? "Last name is required" : "",
    username:
      touched.username && !registerForm.username ? "Username is required" : "",
    email: touched.email && !registerForm.email ? "Email is required" : "",
    password:
      touched.password && !registerForm.password
        ? "Password is required"
        : touched.password && !checkPasswordFormat(registerForm.password)
          ? "Password must be at least character 8, contain a number and a symbol"
          : "",
  };
});

// Check if any field is dirty (modified)
const isDirty = computed(() => {
  return (
    registerForm.firstName !== "" ||
    registerForm.lastName !== "" ||
    registerForm.username !== "" ||
    registerForm.email !== "" ||
    registerForm.password !== ""
  );
});
</script>

<template>
  <div class="page-wrapper">
    <div class="container">
      <div class="container-header">
        <div
          class="tab"
          :class="{ active: activeTab === 'signup' }"
          @click="activeTab = 'signup'"
        >
          Sign-Up
        </div>
        <div
          class="tab"
          :class="{ active: activeTab === 'signin' }"
          @click="activeTab = 'signin'"
        >
          Sign-In
        </div>
      </div>

      <div class="container-body">
        <transition name="fade" mode="out-in">
          <!-- Sign-Up Form -->
          <div
            v-if="activeTab === 'signup'"
            key="signup"
            class="form-container"
          >
            <div class="form-group">
              <label>First Name <span class="required">*</span></label>
              <input
                v-model="registerForm.firstName"
                type="text"
                placeholder="First Name"
                @blur="markTouched('firstName')"
              />
              <p v-if="errors.firstName" class="text-red-600">
                {{ errors.firstName }}
              </p>
            </div>

            <div class="form-group">
              <label>Last Name <span class="required">*</span></label>
              <input
                v-model="registerForm.lastName"
                type="text"
                placeholder="Last Name"
                @blur="markTouched('lastName')"
              />
              <p v-if="errors.lastName" class="text-red-600">
                {{ errors.lastName }}
              </p>
            </div>

            <div class="form-group">
              <label>Username <span class="required">*</span></label>
              <input
                v-model="registerForm.username"
                type="text"
                placeholder="Username"
                @blur="markTouched('username')"
              />
              <p v-if="errors.username" class="text-red-600">
                {{ errors.username }}
              </p>
            </div>

            <div class="form-group">
              <label>Email <span class="required">*</span></label>
              <input
                v-model="registerForm.email"
                type="email"
                placeholder="Email"
                @blur="markTouched('email')"
              />
              <p v-if="errors.email" class="text-red-600">{{ errors.email }}</p>
            </div>

            <div class="form-group">
              <label>Password <span class="required">*</span></label>
              <input
                v-model="registerForm.password"
                type="password"
                placeholder="Password"
                @blur="markTouched('password')"
              />
              <p v-if="errors.password" class="text-red-600">
                {{ errors.password }}
              </p>
            </div>

            <button :disabled="!isDirty" @click.prevent="register">
              Sign Up
            </button>
          </div>

          <!-- Sign-In Form (unchanged) -->
          <div v-else key="signin" class="form-container">
            <div class="form-group">
              <label>Identifier <span class="required">*</span></label>
              <input
                v-model="loginForm.identifier"
                type="text"
                placeholder="Email or Username"
                required
              />
            </div>

            <div class="form-group">
              <label>Password <span class="required">*</span></label>
              <input
                v-model="loginForm.password"
                type="password"
                placeholder="Password"
                required
              />
            </div>

            <button @click.prevent="login">Sign In</button>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<style lang="css" scoped>
.page-wrapper {
  width: 100vw;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #1a202c;
}

.container {
  width: 600px;
  border-radius: 10px;
  border: 0.2px solid gray;
  overflow: hidden;
  background-color: #202c3c;
}

.container-header {
  display: flex;
}

.tab {
  flex: 1;
  text-align: center;
  padding: 14px 0;
  cursor: pointer;
  background-color: #1b2433;
  color: white;
  transition:
    background-color 0.3s,
    color 0.3s;
}

.tab.active {
  background-color: #3c83f6;
  color: white;
}

.container-body {
  padding: 20px;
}

.form-container {
  display: flex;
  flex-direction: column;
}

.form-group {
  display: flex;
  flex-direction: column;
  margin-bottom: 16px;
}

label {
  color: white;
  font-weight: 500;
  margin-bottom: 4px;
}

.required {
  color: red;
  margin-left: 2px;
}

input {
  padding: 10px;
  border-radius: 6px;
  border: none;
  background-color: #1b2433;
  color: white;
}

input::placeholder {
  color: #a0aec0;
}

button {
  padding: 10px;
  border-radius: 6px;
  border: none;
  background-color: #3c83f6;
  color: white;
  cursor: pointer;
}

button:hover {
  background-color: #3370d6;
}

/* Fade transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.4s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.fade-enter-to,
.fade-leave-from {
  opacity: 1;
}
</style>
