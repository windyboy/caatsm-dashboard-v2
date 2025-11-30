import adapter from "@deno/svelte-adapter";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Svelte 5 + vite-plugin-svelte v5: Explicitly enable script preprocessing
  // This ensures TypeScript/JavaScript preprocessing works correctly
  preprocess: vitePreprocess({
    script: true, // Explicitly enable script preprocessing
  }),

  kit: {
    adapter: adapter(),
  },
};

export default config;
