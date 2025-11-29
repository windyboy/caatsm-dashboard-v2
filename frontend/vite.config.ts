/// <reference path="./vite-env.d.ts" />
import devtoolsJson from "vite-plugin-devtools-json";
import { sveltekit } from "@sveltejs/kit/vite";
import UnoCSS from "unocss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [UnoCSS(), sveltekit(), devtoolsJson()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:3002",
        changeOrigin: true,
      },
      "/ws": { target: "ws://localhost:3002", ws: true },
    },
  },
});
