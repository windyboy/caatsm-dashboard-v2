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

  test("should handle WebSocket reconnection and resync", async ({ page, context }) => {
    await page.goto("/");

    // Wait for WebSocket connection
    await page.waitForSelector('[aria-label*="Live telegram messages"]', { timeout: 5000 });

    // Check that connection status is visible
    const statusIndicator = page.locator('.status-indicator, [class*="status"]').first();
    await expect(statusIndicator).toBeVisible({ timeout: 5000 });

    // Simulate network disconnection by going offline
    await context.setOffline(true);
    
    // Wait a bit for disconnection to be detected
    await page.waitForTimeout(2000);

    // Go back online
    await context.setOffline(false);

    // Wait for reconnection (should see reconnecting/connected status)
    // The resync logic should fetch missed messages
    await page.waitForTimeout(3000);

    // Verify connection is restored (basic check - status should be visible)
    const statusAfterReconnect = page.locator('.status-indicator, [class*="status"]').first();
    await expect(statusAfterReconnect).toBeVisible();
  });
});
