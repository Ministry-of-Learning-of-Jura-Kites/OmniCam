export type THEME_OPTION = "light" | "dark";

export function useLightDarkTheme() {
  const theme = ref<"light" | "dark">("light");

  function applyTheme() {
    if (theme.value === "dark") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
    localStorage.setItem("theme", theme.value);
  }

  function toggleTheme() {
    theme.value = theme.value === "light" ? "dark" : "light";
    applyTheme();
  }

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

  return {
    theme,
    toggleTheme,
  };
}
