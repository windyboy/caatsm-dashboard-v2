import { test, expect } from "@playwright/test";

test("Dashboard: should load dashboard page", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveTitle(/CAATSM Dashboard/);
});

test("Dashboard: should display main content", async ({ page }) => {
  await page.goto("/");
  // Just verify page loads without errors
  await expect(page.locator("main")).toBeVisible();
});
