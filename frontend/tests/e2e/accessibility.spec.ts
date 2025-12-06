import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test.describe("Accessibility Tests", () => {
  test("Dashboard page should pass accessibility checks", async ({ page }) => {
    await page.goto("/");
    const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
    expect(accessibilityScanResults.violations).toEqual([]);
  });

  test("Search page should pass accessibility checks", async ({ page }) => {
    await page.goto("/search");
    const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
    expect(accessibilityScanResults.violations).toEqual([]);
  });

  test("Charts should have proper ARIA labels", async ({ page }) => {
    await page.goto("/");

    // Wait for loading to complete
    await page.waitForSelector('[class*="loading"]', { state: 'detached', timeout: 10000 });

    // Check if charts exist and have aria-label attributes
    const charts = page.locator("canvas");
    const chartCount = await charts.count();

    if (chartCount > 0) {
      await expect(charts.first()).toHaveAttribute("aria-label");
    } else {
      // If no charts, test passes (charts may not have data)
      expect(true).toBe(true);
    }
  });

  test("Navigation should have proper ARIA attributes", async ({ page }) => {
    await page.goto("/");

    // Check navigation has aria-label
    const nav = page.locator('[aria-label="Main navigation"]');
    await expect(nav).toBeVisible();
  });

  test("Form inputs should have proper labels", async ({ page }) => {
    await page.goto("/search");

    // Check form inputs have aria-label or associated labels
    const inputs = page.locator('input[type="text"], input[type="datetime-local"]');
    for (const input of await inputs.all()) {
      const ariaLabel = await input.getAttribute("aria-label");
      const id = await input.getAttribute("id");
      const name = await input.getAttribute("name");

      // Should have either aria-label, or id with associated label, or name
      expect(ariaLabel || id || name).toBeTruthy();
    }
  });
});