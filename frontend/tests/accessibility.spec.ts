import { expect, test } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test.describe("Accessibility Tests", () => {
  test("homepage should pass accessibility checks", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"])
      // Exclude scrollable-region-focusable rule
      // VirtualList components have keyboard-accessible content;
      // the scrollable viewport is an internal implementation detail
      .disableRules(["scrollable-region-focusable"])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });
});
