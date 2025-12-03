// Theme store for dark/light mode management

import { writable } from "svelte/store";

export type Theme = "light" | "dark";

const THEME_STORAGE_KEY = "caatsm-theme";

function getInitialTheme(): Theme {
  if (typeof window === "undefined") {
    return "light"; // SSR default
  }

  // Check localStorage first
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  if (stored === "dark" || stored === "light") {
    return stored;
  }

  // Fall back to system preference
  if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
    return "dark";
  }

  return "light";
}

function createThemeStore() {
  const { subscribe, set, update } = writable<Theme>(getInitialTheme());

  // Apply theme to document
  function applyTheme(theme: Theme) {
    if (typeof document === "undefined") return;

    const root = document.documentElement;
    if (theme === "dark") {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }

    // Persist to localStorage
    localStorage.setItem(THEME_STORAGE_KEY, theme);
  }

  // Initialize theme on store creation
  if (typeof window !== "undefined") {
    applyTheme(getInitialTheme());
  }

  return {
    subscribe,
    set: (theme: Theme) => {
      set(theme);
      applyTheme(theme);
    },
    toggle: () => {
      update((current) => {
        const newTheme = current === "dark" ? "light" : "dark";
        applyTheme(newTheme);
        return newTheme;
      });
    },
  };
}

export const theme = createThemeStore();

