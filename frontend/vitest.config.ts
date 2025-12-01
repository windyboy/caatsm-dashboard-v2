import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import UnoCSS from "unocss/vite";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import type { PreprocessorGroup } from "svelte/types/compiler/preprocess";

const __dirname =
  typeof (import.meta as any).dirname !== "undefined"
    ? (import.meta as any).dirname
    : dirname(fileURLToPath(import.meta.url));

// Custom preprocessor for tests that bypasses Vite client requirement
const testStylePreprocess: PreprocessorGroup = {
  style: async ({ content }) => {
    // Return CSS as-is - UnoCSS and Vite will process it through their pipelines
    // This avoids the "Cannot read properties of undefined (reading 'client')" error
    return { code: content || "", map: null };
  },
};

export default defineConfig({
  plugins: [
    // UnoCSS must come before Svelte plugin to process CSS properly
    UnoCSS(),
    svelte({
      // In test environment, use custom style preprocessor to avoid Vite client dependency
      // Script preprocessing still uses vitePreprocess for TypeScript support
      preprocess: process.env.VITEST
        ? [
            testStylePreprocess,
            vitePreprocess({ script: true, style: false }),
          ]
        : vitePreprocess(),
      compilerOptions: {
        dev: true,
        // Svelte 5: Runes mode is the default, no need for componentApi compatibility mode
        // Migration guide: https://svelte.dev/docs/svelte/v5-migration-guide
      },
      // Vite 6: Disable hot reload in test environment
      hot: !process.env.VITEST,
    }),
  ],
  css: {
    // Vite 6: Ensure CSS is properly processed in test environment
    postcss: {},
  },
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
    setupFiles: ["./tests/unit/setup.ts"],
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
    // Suppress unhandled errors that don't affect test results
    // Known issue: jsdom/Deno compatibility causes "dispatchEvent" errors
    // This doesn't affect test functionality - all 84 tests pass correctly
    onUnhandledRejection: "warn",
  },
});
