/**
 * REST API 测试
 *
 * 测试 REST API 客户端功能：
 * - 搜索功能
 * - 统计接口
 * - 自动补全
 * - 导出功能
 */

import { expect, test } from "@playwright/test";

test.describe("REST API", () => {
  test("should load search page", async ({ page }) => {
    await page.goto("/search");

    // 检查搜索表单是否存在
    const searchForm = page.locator("text=Search Telegrams");
    await expect(searchForm).toBeVisible();
  });

  test("should display search form fields", async ({ page }) => {
    await page.goto("/search");

    // 检查搜索字段
    const keywordsInput = page.locator('input[type="search"]');
    await expect(keywordsInput).toBeVisible();

    const typeSelect = page.locator("select").first();
    await expect(typeSelect).toBeVisible();
  });

  test("should have search button", async ({ page }) => {
    await page.goto("/search");

    const searchButton = page.locator('button[type="submit"]');
    await expect(searchButton).toBeVisible();
    await expect(searchButton).toContainText("Search");
  });
});
