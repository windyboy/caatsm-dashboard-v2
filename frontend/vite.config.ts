/// <reference path="./vite-env.d.ts" />
import devtoolsJson from "vite-plugin-devtools-json";
import { sveltekit } from "@sveltejs/kit/vite";
import UnoCSS from "unocss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  // Plugin ordering for Svelte 5:
  // 1. UnoCSS first (processes CSS)
  // 2. sveltekit() includes @sveltejs/vite-plugin-svelte (processes Svelte files)
  // 3. devtoolsJson last (development tooling)
  // This order ensures proper HMR behavior with Svelte 5
  plugins: [UnoCSS(), sveltekit(), devtoolsJson()],
  server: {
    port: 5173,
    // HMR settings for Svelte 5 - ensure proper hot module replacement
    hmr: {
      overlay: true, // Show error overlay in development
    },
    proxy: {
      "/api": {
        target: "http://localhost:3002",
        changeOrigin: true,
      },
      "/ws": { target: "ws://localhost:3002", ws: true },
    },
  },
});
