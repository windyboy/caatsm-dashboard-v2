import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname =
  typeof (import.meta as any).dirname !== "undefined"
    ? (import.meta as any).dirname
    : dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [
    svelte({
      preprocess: vitePreprocess(),
      compilerOptions: {
        dev: true,
        // Svelte 5: Remove componentApi compatibility mode
        // Migration guide: https://svelte.dev/docs/svelte/v5-migration-guide
      },
    }),
  ],
  resolve: {
    alias: {
      $lib: resolve(__dirname, "./src/lib"),
    },
    conditions: ["browser", "import"],
  },
  optimizeDeps: {
    include: ["svelte"],
  },
  define: {
    "import.meta.env.DEV": "true",
    "import.meta.env.VITE_API_BASE_URL": '""',
    "import.meta.env.VITE_WS_URL": '""',
    "import.meta.env.VITE_LOG_LEVEL": '""',
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
});
