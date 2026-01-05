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
import { Moon, Sun } from "lucide-vue-next";
import FailDialog from "~/components/dialog/FailDialog.vue";
import { useFailDialog } from "~/composables/useFailDialog";
import { useAuth } from "~/composables/useAuth";
const { open, message } = useFailDialog();

const { theme, toggleTheme } = useLightDarkTheme();
const config = useRuntimeConfig();
const auth = await useAuth();
const { user } = auth;

const handleLogout = () => {
  try {
    $fetch<null>(
      "http://" + config.public.externalBackendHost + "/api/v1/logout",
      {
        method: "POST",
        credentials: "include",
      },
    );
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
              <Moon v-if="theme === 'light'" />
              <Sun v-else />
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
