import { defineConfig, devices } from "@playwright/test";
import process from "node:process";

export default defineConfig({
  testDir: "./tests",
  testMatch: /.*\.(spec|test)\.(ts|js|mjs)/,
  testIgnore: [
    "**/unit/**", // Ignore unit tests (they use vitest, not Playwright)
    "**/setup.ts", // Ignore Vitest setup file
    "**/node_modules/**",
  ],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? [["junit", { outputFile: "test-results.xml" }], ["html"]] : "html",
  use: {
    baseURL: process.env.BASE_URL || "http://localhost:5173",
    trace: "on-first-retry",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "mobile",
      use: { ...devices["Pixel 5"] },
    },
  ],
  webServer: {
    command: process.env.USE_DENO === "true" 
      ? "deno task dev" 
      : (process.env.USE_BUN === "true" ? "bun run dev" : "npm run dev"),
    url: process.env.BASE_URL || "http://localhost:5173",
    reuseExistingServer: !process.env.CI,
    timeout: 60 * 1000,
  },
});
