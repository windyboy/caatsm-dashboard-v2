import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Accessibility Tests', () => {
  test('homepage should pass accessibility checks', async ({ page }) => {
    await page.goto('/');

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });

  test('search page should pass accessibility checks', async ({ page }) => {
    await page.goto('/search');

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });

  test('homepage keyboard navigation', async ({ page }) => {
    await page.goto('/');

    // Test tab navigation
    await page.keyboard.press('Tab');
    let focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    expect(focusedElement).toBeDefined();

    // Continue tabbing through interactive elements
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab');
      await page.waitForTimeout(100); // Small delay for visual feedback
    }

    // Should be able to tab through all interactive elements without getting stuck
    const finalFocusedElement = await page.evaluate(() => document.activeElement?.tagName);
    expect(finalFocusedElement).toBeDefined();
  });

  test('search page keyboard navigation', async ({ page }) => {
    await page.goto('/search');

    // Focus on search input
    const searchInput = page.locator('input[type="search"]').or(
      page.locator('input[placeholder*="search"]')
    );

    if (await searchInput.count() > 0) {
      await searchInput.first().focus();

      // Test that input is focused
      const isFocused = await searchInput.first().evaluate(el => el === document.activeElement);
      expect(isFocused).toBe(true);

      // Test typing in search
      await page.keyboard.type('test query');
      const inputValue = await searchInput.first().inputValue();
      expect(inputValue).toBe('test query');
    } else {
      test.skip();
    }
  });

  test('form elements have proper labels', async ({ page }) => {
    await page.goto('/search');

    // Check all form inputs have labels or aria-labels
    const inputs = page.locator('input, select, textarea');
    const inputCount = await inputs.count();

    for (let i = 0; i < inputCount; i++) {
      const input = inputs.nth(i);
      const hasLabel = await input.evaluate(el => {
        const id = el.id;
        const ariaLabel = el.getAttribute('aria-label');
        const ariaLabelledBy = el.getAttribute('aria-labelledby');
        const label = id ? document.querySelector(`label[for="${id}"]`) : null;
        return !!(ariaLabel || ariaLabelledBy || label);
      });

      expect(hasLabel).toBe(true);
    }
  });

  test('buttons have accessible names', async ({ page }) => {
    await page.goto('/');

    const buttons = page.locator('button, [role="button"]');
    const buttonCount = await buttons.count();

    for (let i = 0; i < buttonCount; i++) {
      const button = buttons.nth(i);
      const accessibleName = await button.evaluate(el => {
        const text = el.textContent?.trim();
        const ariaLabel = el.getAttribute('aria-label');
        const title = el.getAttribute('title');
        return text || ariaLabel || title;
      });

      expect(accessibleName).toBeTruthy();
    }
  });

  test('images have alt text', async ({ page }) => {
    await page.goto('/');

    const images = page.locator('img');
    const imageCount = await images.count();

    for (let i = 0; i < imageCount; i++) {
      const image = images.nth(i);
      const altText = await image.getAttribute('alt');
      const ariaHidden = await image.getAttribute('aria-hidden');

      // Images should have alt text unless they are decorative (aria-hidden)
      if (ariaHidden !== 'true') {
        expect(altText).toBeTruthy();
      }
    }
  });

  test('color contrast meets WCAG standards', async ({ page }) => {
    await page.goto('/');

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .withRules(['color-contrast'])
      .analyze();

    // Filter for color contrast violations only
    const colorContrastViolations = accessibilityScanResults.violations.filter(
      violation => violation.id === 'color-contrast'
    );

    expect(colorContrastViolations).toEqual([]);
  });

  test('focus indicators are visible', async ({ page }) => {
    await page.goto('/');

    // Focus on first focusable element
    await page.keyboard.press('Tab');

    // Check if focus indicator is visible (this is a basic check)
    const focusedElement = await page.evaluate(() => {
      const el = document.activeElement;
      if (!el) return false;

      const style = window.getComputedStyle(el);
      return style.outline !== 'none' ||
             style.boxShadow !== 'none' ||
             el.classList.contains('focus-visible') ||
             el.classList.contains('ring');
    });

    expect(focusedElement).toBe(true);
  });

  test('page has proper heading structure', async ({ page }) => {
    await page.goto('/');

    // Check for h1 tag
    const h1Count = await page.locator('h1').count();
    expect(h1Count).toBeGreaterThan(0);

    // Check heading hierarchy (basic check)
    const headings = await page.locator('h1, h2, h3, h4, h5, h6').allTextContents();
    expect(headings.length).toBeGreaterThan(0);
  });

  test('mobile accessibility', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });
});