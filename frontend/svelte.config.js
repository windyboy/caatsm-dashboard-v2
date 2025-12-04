import adapter from "@deno/svelte-adapter";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Svelte 5 + vite-plugin-svelte v5: Explicitly enable script preprocessing
  // This ensures TypeScript/JavaScript preprocessing works correctly
  preprocess: vitePreprocess({
    script: true, // Explicitly enable script preprocessing
  }),

  // Suppress CSS unused selector warnings for dynamically added attributes
  // These selectors are used by Melt UI and other libraries at runtime
  onwarn: (warning, handler) => {
    // Suppress CSS unused selector warnings - these are false positives
    // Selectors targeting dynamically added attributes (like data-state, data-highlighted)
    // are used by libraries like Melt UI at runtime
    if (warning.code === "css-unused-selector") {
      return;
    }
    handler(warning);
  },

  kit: {
    adapter: adapter(),
  },
};

export default config;
