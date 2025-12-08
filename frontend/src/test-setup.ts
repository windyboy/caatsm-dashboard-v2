import { afterEach, beforeAll, vi } from "vitest";
import { cleanup } from "@testing-library/svelte";
import { init, register } from "svelte-i18n";

// Initialize i18n for tests
beforeAll(async () => {
  register("en", () => import("$lib/i18n/locales/en.json"));
  register("zh", () => import("$lib/i18n/locales/zh.json"));
  register("zh-CN", () => import("$lib/i18n/locales/zh.json"));
  await init({
    fallbackLocale: "en",
    initialLocale: "en",
  });
});

// Cleanup after each test
afterEach(() => {
  cleanup();
});
