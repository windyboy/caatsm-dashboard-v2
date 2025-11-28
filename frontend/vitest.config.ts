import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname =
  typeof import.meta.dirname !== "undefined"
    ? import.meta.dirname
    : dirname(fileURLToPath(import.meta.url));

// Workaround for Deno compatibility - create a minimal plugin that doesn't use hot updates
const sveltePlugin = svelte({
  hot: false,
  compilerOptions: {
    dev: false,
  },
});

// Override the configureServer hook to prevent the error
const safeSveltePlugin = {
  ...sveltePlugin,
  configureServer(server: any) {
    // Only call configureServer if it exists and server.ws is available
    if (sveltePlugin.configureServer && server?.ws) {
      return sveltePlugin.configureServer(server);
    }
  },
};

export default defineConfig({
  plugins: [safeSveltePlugin],
  server: {
    ws: {},
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./tests/setup.ts"],
    include: ["tests/unit/**/*.test.ts"],
    exclude: [
      "node_modules",
      "tests/**/*.spec.ts",
      "tests/api.test.ts",
      "tests/websocket.test.ts",
    ],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: ["node_modules/", "tests/", "**/*.config.*", "**/*.d.ts"],
    },
  },
  resolve: {
    alias: {
      $lib: resolve(__dirname, "./src/lib"),
    },
  },
});
