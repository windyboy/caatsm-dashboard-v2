import { register, init, getLocaleFromNavigator } from "svelte-i18n";
import { browser } from "$app/environment";

// Locale normalization function
// Ensures consistent locale handling across server and client
function normalizeLocale(locale: string | null | undefined): string {
  if (!locale) return "en";
  const lower = locale.toLowerCase();
  if (lower.startsWith("zh")) return "zh-CN"; // Default to simplified Chinese
  if (lower.startsWith("en")) return "en";
  return "en"; // fallback
}

// Only execute on client side
// Server side initialization is handled by hooks.server.ts
if (browser) {
  try {
    // Register locales with lazy loading (recommended by svelte-i18n)
    // Register multiple locale keys to support browser standard formats
    register("en", () => import("./locales/en.json"));
    register("en-US", () => import("./locales/en.json")); // Compatible with browser
    register("zh", () => import("./locales/zh.json"));
    register("zh-CN", () => import("./locales/zh.json")); // Compatible with browser

    // Get locale from cookie (set by server), navigator, or fallback to 'en'
    function getInitialLocale(): string {
      // Try to read locale from cookie (set by server)
      const cookieLocale = document.cookie
        .split("; ")
        .find((row) => row.startsWith("locale="))
        ?.split("=")[1];

      if (cookieLocale) {
        return normalizeLocale(cookieLocale);
      }

      // Fallback to navigator locale or 'en'
      const navigatorLocale = getLocaleFromNavigator();
      return normalizeLocale(navigatorLocale);
    }

    // Initialize i18n synchronously
    // CRITICAL: This MUST run before any component uses $t() or formatMessage()
    // The fallbackLocale ensures we always have a valid locale, preventing hydration errors
    // init() is idempotent and safe to call multiple times
    init({
      fallbackLocale: "en",
      initialLocale: getInitialLocale(),
    });
  } catch (error) {
    // Silently fail on client side if there's an error
    // Server side will handle initialization via hooks.server.ts
    console.warn("Failed to initialize i18n on client side:", error);
  }
}
