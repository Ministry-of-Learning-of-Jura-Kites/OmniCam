export type THEME_OPTION = "light" | "dark";

export function useLightDarkTheme() {
  const theme = useCookie<THEME_OPTION | undefined>("theme", {
    default: () => "light" as const,
  });

  function toggleTheme() {
    theme.value = theme.value === "light" ? "dark" : "light";
  }

  return {
    theme,
    toggleTheme,
  };
}
