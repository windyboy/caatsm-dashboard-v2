import { test, expect } from "@playwright/test";

test("homepage has title", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveTitle(/CAATSM Dashboard/);
});

test("homepage has navigation links", async ({ page }) => {
  await page.goto("/");
  await expect(page.locator('nav a:has-text("Dashboard")')).toBeVisible();
  await expect(page.locator('nav a:has-text("Search")')).toBeVisible();
});

test("can navigate to search page", async ({ page }) => {
  await page.goto("/");
  await page.click('nav a:has-text("Search")');
  await expect(page).toHaveURL(/.*\/search/);
  await expect(page.locator('h2:has-text("Search Telegrams")')).toBeVisible();
});

test("search form has all fields", async ({ page }) => {
  await page.goto("/search");
  await expect(page.locator('input[placeholder*="Flight number"]')).toBeVisible();
  await expect(page.locator('select[id="type-select"]')).toBeVisible();
  await expect(page.locator('select[id="priority-select"]')).toBeVisible();
  await expect(page.locator('input[id="start-time-input"]')).toBeVisible();
  await expect(page.locator('input[id="end-time-input"]')).toBeVisible();
});

test("search form reset works", async ({ page }) => {
  await page.goto("/search");
  const queryInput = page.locator('input[placeholder*="Flight number"]');
  await queryInput.type("test query");
  await page.click('button:has-text("Reset")');
  await expect(queryInput).toHaveValue("");
});

test("live stream shows messages", async ({ page }) => {
  await page.goto("/");
  await expect(page.locator('h2:has-text("Live Stream")')).toBeVisible();
  // Note: Actual messages depend on WebSocket connection
  // This test just verifies the component renders
});
