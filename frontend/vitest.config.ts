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
      // Svelte 5 + vite-plugin-svelte v5: Explicitly enable script preprocessing
      // Vite 6: In test environment, skip CSS preprocessing to avoid client dependency issues
      preprocess: vitePreprocess({
        script: true, // Explicitly enable script preprocessing for TypeScript/JavaScript
        style: !process.env.VITEST, // Disable CSS preprocessing in test environment (Vite 6 compatibility)
      }),
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
