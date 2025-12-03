// Theme store for dark/light mode management

import { writable } from "svelte/store";
import { isBrowser, getWindow, getDocument } from "../utils/browser";

export type Theme = "light" | "dark";

const THEME_STORAGE_KEY = "caatsm-theme";

function getInitialTheme(): Theme {
  if (!isBrowser) {
    return "light"; // SSR default
  }

  const win = getWindow();
  if (!win) return "light";

  // Check localStorage first
  const stored = win.localStorage.getItem(THEME_STORAGE_KEY);
  if (stored === "dark" || stored === "light") {
    return stored;
  }

  // Fall back to system preference
  if (win.matchMedia("(prefers-color-scheme: dark)").matches) {
    return "dark";
  }

  return "light";
}

function createThemeStore() {
  const { subscribe, set, update } = writable<Theme>(getInitialTheme());

  // Apply theme to document
  function applyTheme(theme: Theme) {
    const doc = getDocument();
    if (!doc) return;

    const win = getWindow();
    if (!win) return;

    const root = doc.documentElement;
    if (theme === "dark") {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }

    // Persist to localStorage
    win.localStorage.setItem(THEME_STORAGE_KEY, theme);
  }

  // Initialize theme on store creation
  if (isBrowser) {
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

