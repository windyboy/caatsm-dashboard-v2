#!/usr/bin/env -S deno run -A
/**
 * Wrapper script for running Vite dev server in Deno
 * Sets USE_DENO=true to disable HMR and avoid WebSocket broken pipe errors
 */

// Set environment variable for Vite config
Deno.env.set("USE_DENO", "true");

// Run Vite dev server
const vite = Deno.run({
  cmd: ["deno", "run", "-A", "npm:vite@7.2.6", "dev"],
  env: {
    ...Deno.env.toObject(),
    USE_DENO: "true",
  },
});

const status = await vite.status();
Deno.exit(status.code);

