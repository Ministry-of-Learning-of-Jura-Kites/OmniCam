<!-- layouts/default.vue -->
<script setup lang="ts">
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import FailDialog from "~/components/dialog/FailDialog.vue";
import { useFailDialog } from "~/composables/useFailDialog";
import { useAuth } from "~/composables/useAuth";
const { open, message } = useFailDialog();

const theme = ref<"light" | "dark">("light");
const config = useRuntimeConfig();
const auth = await useAuth();
const { user } = auth;
onMounted(() => {
  const savedTheme = localStorage.getItem("theme");
  const systemPrefersDark = window.matchMedia(
    "(prefers-color-scheme: dark)",
  ).matches;

  if (savedTheme) {
    theme.value = savedTheme as "light" | "dark";
  } else if (systemPrefersDark) {
    theme.value = "dark";
  }

  applyTheme();
});

const applyTheme = () => {
  if (theme.value === "dark") {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
  localStorage.setItem("theme", theme.value);
};

const toggleTheme = () => {
  theme.value = theme.value === "light" ? "dark" : "light";
  applyTheme();
};

const handleLogout = () => {
  try {
    $fetch<null>("http://" + config.public.backendHost + "/api/v1/logout", {
      method: "POST",
      credentials: "include",
    });
    navigateTo("/authentication");
  } catch (err) {
    console.log(err);
  }
};
</script>
<template>
  <div class="min-h-screen bg-white dark:bg-gray-900">
    <!-- Header -->
    <header
      class="sticky top-0 z-50 border-b bg-white dark:bg-gray-800 dark:border-gray-700"
    >
      <div class="container mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <!-- Logo -->
          <div class="flex items-center">
            <NuxtLink
              to="/"
              class="text-xl font-bold text-gray-900 dark:text-white"
            >
              Omnicam
            </NuxtLink>
          </div>

          <!-- Right Section -->
          <div class="flex items-center space-x-4">
            <button
              class="p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              aria-label="Toggle theme"
              @click="toggleTheme"
            >
              <svg
                v-if="theme === 'light'"
                class="w-5 h-5 text-gray-600 dark:text-gray-300"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
                />
              </svg>
              <svg
                v-else
                class="w-5 h-5 text-gray-600 dark:text-gray-300"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
                />
              </svg>
            </button>

            <!-- Profile Dropdown -->
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <button class="flex items-center space-x-2 focus:outline-none">
                  <div
                    class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center"
                  >
                    <span class="text-white text-sm font-medium">
                      {{ user?.username[0]?.toUpperCase() ?? "?" }}
                    </span>
                  </div>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent class="w-56">
                <DropdownMenuLabel>{{ user?.first_name }}</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem>
                    <NuxtLink to="/profile" class="w-full"> Profile </NuxtLink>
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <NuxtLink to="/settings" class="w-full">
                      Settings
                    </NuxtLink>
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuItem @click="handleLogout">
                  Log out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="bg-white dark:bg-gray-900">
      <slot />
    </main>

    <FailDialog
      v-model:open="open"
      :message="message"
      icon="fa-solid fa-circle-exclamation"
    />
  </div>
</template>
