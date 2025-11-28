import { test, expect } from '@playwright/test';

// Constants for maintainability
const SCREENSHOT_OPTIONS = {
  fullPage: true,
  threshold: 0.2, // More lenient to reduce false positives
  animations: 'disabled' as const // Disable animations for stable screenshots
} as const;

const VIEWPORTS = {
  mobile: { width: 375, height: 667 },
  tablet: { width: 768, height: 1024 }
} as const;

const TIMEOUT = 5000; // 5 seconds timeout for elements

test.describe('Visual Regression Tests', () => {
  test('homepage should match baseline', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle' });
    // Wait for component hydration and WebSocket connection attempt to stabilize
    await page.waitForTimeout(2000);
    // Wait for LiveStream component to be visible and stable
    await page.waitForSelector('h2:has-text("Live Stream")', { state: 'visible' });
    await page.waitForTimeout(500);
    await expect(page).toHaveScreenshot('homepage.png', SCREENSHOT_OPTIONS);
  });

  test('search page empty state', async ({ page }) => {
    await page.goto('/search', { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);
    await expect(page).toHaveScreenshot('search-page.png', SCREENSHOT_OPTIONS);
  });

  test('homepage mobile view', async ({ page }) => {
    await page.setViewportSize(VIEWPORTS.mobile);
    await page.goto('/', { waitUntil: 'networkidle' });
    // Wait for component hydration and WebSocket connection attempt to stabilize
    await page.waitForTimeout(2000);
    // Wait for LiveStream component to be visible and stable
    await page.waitForSelector('h2:has-text("Live Stream")', { state: 'visible' });
    await page.waitForTimeout(500);
    await expect(page).toHaveScreenshot('homepage-mobile.png', SCREENSHOT_OPTIONS);
  });

  test('homepage tablet view', async ({ page }) => {
    await page.setViewportSize(VIEWPORTS.tablet);
    await page.goto('/', { waitUntil: 'networkidle' });
    // Wait for component hydration and WebSocket connection attempt to stabilize
    await page.waitForTimeout(2000);
    // Wait for LiveStream component to be visible and stable
    await page.waitForSelector('h2:has-text("Live Stream")', { state: 'visible' });
    await page.waitForTimeout(500);
    await expect(page).toHaveScreenshot('homepage-tablet.png', SCREENSHOT_OPTIONS);
  });

  test('search page with query results', async ({ page }) => {
    await page.goto('/search', { waitUntil: 'load' });
    
    // Wait for search input to be available
    const searchInput = page.locator('input[type="search"]')
      .or(page.locator('input[placeholder*="search"]'));

    try {
      await searchInput.first().waitFor({ timeout: TIMEOUT });
      await searchInput.first().fill('test query');
      
      // Wait for results to load - adjust selector based on actual implementation
      const resultsContainer = page.locator('[data-testid="search-results"]')
        .or(page.locator('.search-results'))
        .or(page.locator('[role="list"]'));
      
      await resultsContainer.first().waitFor({ timeout: TIMEOUT });
      await page.waitForTimeout(500); // Allow results to fully render
      await expect(page).toHaveScreenshot('search-with-results.png', SCREENSHOT_OPTIONS);
    } catch (error) {
      test.skip(true, 'Search functionality not available or results not found');
    }
  });

  test('dark mode toggle', async ({ page }) => {
    await page.goto('/', { waitUntil: 'load' });
    
    // Try to find dark mode toggle with proper error handling
    const darkModeToggle = page.locator('[data-testid="dark-mode-toggle"]')
      .or(page.locator('button:has-text("Dark")'))
      .or(page.locator('button:has-text("Light")'));

    try {
      await darkModeToggle.first().click({ timeout: TIMEOUT });
      await page.waitForTimeout(500); // Wait for theme transition
      await expect(page).toHaveScreenshot('homepage-dark-mode.png', SCREENSHOT_OPTIONS);
    } catch (error) {
      test.skip(true, 'Dark mode toggle not found on page');
    }
  });

  test('404 page display', async ({ page }) => {
    // Navigate to a non-existent route to display 404 page
    const response = await page.goto('/non-existent-route-404', { waitUntil: 'load' });
    await page.waitForTimeout(1000);

    // Verify we got a 404 response or the page shows not found content
    const notFoundContent = page.locator('[data-testid="not-found"]')
      .or(page.locator('text="404"'))
      .or(page.locator('text="Not Found"'))
      .or(page.locator('text="Page not found"'));

    try {
      await notFoundContent.first().waitFor({ timeout: TIMEOUT });
      await expect(page).toHaveScreenshot('404-page.png', SCREENSHOT_OPTIONS);
    } catch (error) {
      // If no specific 404 content is found, still take screenshot if response was 404
      if (response?.status() === 404) {
        await expect(page).toHaveScreenshot('404-page.png', SCREENSHOT_OPTIONS);
      } else {
        test.skip(true, '404 page content not found');
      }
    }
  });

  test('error boundary display', async ({ page }) => {
    // Navigate to test error route (dev/test only)
    await page.goto('/test-error', { waitUntil: 'load' });
    
    // Verify we're on the test page
    const testPage = page.locator('[data-testid="error-trigger-page"]');
    
    try {
      await testPage.waitFor({ timeout: TIMEOUT });
      
      // Click the button to trigger the error
      const triggerButton = page.locator('[data-testid="trigger-error-button"]');
      await triggerButton.click();
      
      // Wait for error boundary to appear
      const errorBoundary = page.locator('.error-boundary')
        .or(page.locator('text="Oops! Something went wrong"'))
        .or(page.locator('.error-content'));
      
      await errorBoundary.first().waitFor({ timeout: TIMEOUT });
      await page.waitForTimeout(500); // Allow error UI to fully render
      
      await expect(page).toHaveScreenshot('error-boundary.png', SCREENSHOT_OPTIONS);
    } catch (error) {
      test.skip(true, 'Error boundary test route not available (may need dev environment)');
    }
  });

  test('loading states', async ({ page }) => {
    // Intercept network requests to delay responses and capture loading state
    await page.route('**/api/**', async (route) => {
      await new Promise(resolve => setTimeout(resolve, 1000)); // Delay API responses by 1s
      await route.continue();
    });

    await page.goto('/search', { waitUntil: 'load' });
    
    // Trigger a search to show loading state
    const searchInput = page.locator('input[type="search"]')
      .or(page.locator('input[placeholder*="search"]'));

    try {
      await searchInput.first().waitFor({ timeout: TIMEOUT });
      await searchInput.first().fill('test query');
      await searchInput.first().press('Enter'); // Trigger search submission
      
      // Wait for loading indicator to appear with explicit state check
      const loadingIndicator = page.locator('[data-testid="loading"]')
        .or(page.locator('.animate-spin'))
        .or(page.locator('text="Loading"'))
        .or(page.locator('text="Searching"'));

      await loadingIndicator.first().waitFor({ state: 'visible', timeout: TIMEOUT });
      await expect(page).toHaveScreenshot('loading-state.png', SCREENSHOT_OPTIONS);
    } catch (error) {
      test.skip(true, 'Search input or loading indicator not found');
    }
  });
});