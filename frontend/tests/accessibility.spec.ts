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
    const focusedElement = await page.evaluate(() => {
      const el = document.activeElement;
      return el && el !== document.body && el !== document.documentElement ? el.tagName : null;
    });
    expect(focusedElement).not.toBeNull();

    // Continue tabbing through interactive elements
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab');
    }

    // Should be able to tab through all interactive elements without getting stuck
    const finalFocusedElement = await page.evaluate(() => {
      const el = document.activeElement;
      return el && el !== document.body && el !== document.documentElement ? el.tagName : null;
    });
    expect(finalFocusedElement).not.toBeNull();
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
    // Disable animations and transitions to prevent flakiness
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.addStyleTag({
      content: `
        *, *::before, *::after {
          animation-duration: 0s !important;
          animation-delay: 0s !important;
          transition-duration: 0s !important;
          transition-delay: 0s !important;
        }
      `
    });

    await page.goto('/');
    
    // Wait for page to be fully loaded and stable
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(100); // Small buffer for any final renders

    // Get the first focusable element
    const firstFocusable = await page.locator(
      'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
    ).first();
    
    // Ensure element exists
    await expect(firstFocusable).toBeVisible();

    // Get styles before focus
    const unfocusedStyles = await firstFocusable.evaluate(el => {
      const computed = globalThis.getComputedStyle(el);
      return {
        outline: computed.outline,
        outlineWidth: computed.outlineWidth,
        outlineColor: computed.outlineColor,
        borderColor: computed.borderColor,
        boxShadow: computed.boxShadow
      };
    });

    // Focus the element
    await page.keyboard.press('Tab');
    await page.waitForTimeout(50); // Small delay for focus styles to apply

    // Verify element is focused
    const isFocused = await firstFocusable.evaluate(el => el === document.activeElement);
    expect(isFocused).toBe(true);

    // Get styles after focus
    const focusedStyles = await firstFocusable.evaluate(el => {
      const computed = globalThis.getComputedStyle(el);
      return {
        outline: computed.outline,
        outlineWidth: computed.outlineWidth,
        outlineColor: computed.outlineColor,
        borderColor: computed.borderColor,
        boxShadow: computed.boxShadow
      };
    });

    // Verify at least one focus style property changed
    const hasVisibleFocusIndicator = 
      unfocusedStyles.outline !== focusedStyles.outline ||
      unfocusedStyles.outlineWidth !== focusedStyles.outlineWidth ||
      unfocusedStyles.outlineColor !== focusedStyles.outlineColor ||
      unfocusedStyles.borderColor !== focusedStyles.borderColor ||
      unfocusedStyles.boxShadow !== focusedStyles.boxShadow;
    
    expect(hasVisibleFocusIndicator).toBe(true);
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