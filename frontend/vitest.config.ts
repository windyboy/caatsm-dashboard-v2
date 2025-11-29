import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
// @ts-expect-error - node:path is available at runtime
import { dirname, resolve } from "node:path";
// @ts-expect-error - node:url is available at runtime
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
        compatibility: {
          componentApi: 4,
        },
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
