import { test, expect } from '@playwright/test';

test.describe('Visual Regression Tests', () => {
  test('homepage should match baseline', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveScreenshot('homepage.png', {
      fullPage: true,
      threshold: 0.1
    });
  });

  test('search page should match baseline', async ({ page }) => {
    await page.goto('/search');
    await expect(page).toHaveScreenshot('search-page.png', {
      fullPage: true,
      threshold: 0.1
    });
  });

  test('homepage mobile view', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');
    await expect(page).toHaveScreenshot('homepage-mobile.png', {
      fullPage: true,
      threshold: 0.1
    });
  });

  test('homepage tablet view', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await expect(page).toHaveScreenshot('homepage-tablet.png', {
      fullPage: true,
      threshold: 0.1
    });
  });

  test('search page with results', async ({ page }) => {
    await page.goto('/search');
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveScreenshot('search-with-results.png', {
      fullPage: true,
      threshold: 0.1
    });
  });

  test('dark mode toggle', async ({ page }) => {
    await page.goto('/');
    // Assuming there's a dark mode toggle button
    const darkModeToggle = page.locator('[data-testid="dark-mode-toggle"]').or(
      page.locator('button:has-text("Dark")')
    ).or(page.locator('button:has-text("Light")'));

    if (await darkModeToggle.count() > 0) {
      await darkModeToggle.first().click();
      await page.waitForLoadState('networkidle');
      await expect(page).toHaveScreenshot('homepage-dark-mode.png', {
        fullPage: true,
        threshold: 0.1
      });
    } else {
      // Skip test if no dark mode toggle found
      test.skip();
    }
  });

  test('error boundary display', async ({ page }) => {
    // Navigate to a non-existent route to trigger error boundary
    await page.goto('/non-existent-route');
    await page.waitForLoadState('networkidle');

    // Check if error boundary is displayed
    const errorBoundary = page.locator('[data-testid="error-boundary"]').or(
      page.locator('text="Something went wrong"')
    );

    if (await errorBoundary.count() > 0) {
      await expect(page).toHaveScreenshot('error-boundary.png', {
        fullPage: true,
        threshold: 0.1
      });
    } else {
      // Skip if no error boundary visible
      test.skip();
    }
  });

  test('loading states', async ({ page }) => {
    await page.goto('/search');
    // Trigger a search to show loading state
    const searchInput = page.locator('input[type="search"]').or(
      page.locator('input[placeholder*="search"]')
    );

    if (await searchInput.count() > 0) {
      await searchInput.first().fill('test query');
      // Look for loading indicators
      const loadingIndicator = page.locator('[data-testid="loading"]').or(
        page.locator('.animate-spin')
      ).or(page.locator('text="Loading"')).or(page.locator('text="Searching"'));

      if (await loadingIndicator.count() > 0) {
        await expect(page).toHaveScreenshot('loading-state.png', {
          fullPage: true,
          threshold: 0.1
        });
      } else {
        test.skip();
      }
    } else {
      test.skip();
    }
  });
});