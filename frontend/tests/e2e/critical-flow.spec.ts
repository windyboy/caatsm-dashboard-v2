import { test, expect } from "@playwright/test";

test.describe("Critical User Flows", () => {
  test("should validate search time range", async ({ page }) => {
    await page.goto("/search");

    // Test 90-day limit
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - 100);
    await page.fill("#start-time-input", startDate.toISOString().slice(0, 16));

    const endDate = new Date();
    await page.fill("#end-time-input", endDate.toISOString().slice(0, 16));

    await page.click('button[type="submit"]');
    await expect(page.locator("text=/Time range cannot exceed 90 days/i")).toBeVisible({
      timeout: 5000,
    });
  });
});
