/// <reference path="./vite-env.d.ts" />
import { sveltekit } from "@sveltejs/kit/vite";
import UnoCSS from "unocss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  // Plugin ordering for Svelte 5:
  // 1. UnoCSS first (processes CSS)
  // 2. sveltekit() includes @sveltejs/vite-plugin-svelte (processes Svelte files)
  // This order ensures proper HMR behavior with Svelte 5
  plugins: [UnoCSS(), sveltekit()],
  server: {
    port: 5173,
    // HMR settings for Svelte 5 - ensure proper hot module replacement
    // Note: HMR is disabled when USE_DENO=true to avoid WebSocket broken pipe errors (os error 32)
    // This is a known compatibility issue between Vite 7 and Deno
    hmr:
      process.env.USE_DENO === "true"
        ? false // Disable HMR in Deno to avoid WebSocket broken pipe errors
        : {
            overlay: true, // Show error overlay in development
            clientPort: 5173,
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
