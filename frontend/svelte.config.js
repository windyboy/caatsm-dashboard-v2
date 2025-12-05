import adapter from "@deno/svelte-adapter";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import process from "node:process";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Svelte 5 + vite-plugin-svelte v5: Explicitly enable script preprocessing
  // This ensures TypeScript/JavaScript preprocessing works correctly
  preprocess: vitePreprocess({
    script: true, // Explicitly enable script preprocessing
  }),

  kit: {
    adapter: adapter(),
    // Disable service worker in development to avoid errors
    // Service worker will only be registered in production builds
    serviceWorker: {
      register: process.env.NODE_ENV === "production",
    },
  },
};

export default config;
