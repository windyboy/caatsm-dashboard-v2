import { test, expect } from "@playwright/test";

test("Search: should load search page", async ({ page }) => {
  await page.goto("/search");
  await expect(page).toHaveTitle(/Search/);
});

test("Search: should display search form", async ({ page }) => {
  await page.goto("/search");
  await expect(page.locator('input[name="query"]')).toBeVisible();
});

test("Search: should show error for invalid time range", async ({ page }) => {
  await page.goto("/search");
  
  const startTime = page.locator('input[name="start_time"]');
  const endTime = page.locator('input[name="end_time"]');
  
  await startTime.fill("2024-01-02T00:00");
  await endTime.fill("2024-01-01T00:00"); // End before start
  
  await page.locator('button[type="submit"]').click();
  
  // Should show validation error
  await expect(page.locator('[role="alert"]')).toBeVisible();
});
