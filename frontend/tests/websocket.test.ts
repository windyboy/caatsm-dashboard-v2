/**
 * WebSocket 连接测试
 *
 * 测试 WebSocket 客户端功能：
 * - 连接建立
 * - 消息接收
 * - 自动重连
 * - 统计更新
 */

import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

test.describe("WebSocket Connection", () => {
  test("should load page without WebSocket backend", async ({ page }) => {
    // Note: WebSocket connection requires backend running on port 3002
    // This test verifies the page loads and components mount without requiring WS connection

    await page.goto("/");

    // Wait for the page to be fully loaded (ensures components mounted)
    await expect(page.locator("h1")).toBeVisible();
    await expect(page.locator("h1")).toContainText("CAATSM Dashboard");

    // Verify page loads successfully (WS connection is optional for this test)
  });
  test("should display live stream component", async ({ page }: { page: Page }) => {
    await page.goto("/");

    // 检查 LiveStream 组件是否存在
    const liveStream = page.locator("text=Live Stream");
    await expect(liveStream).toBeVisible();
  });

  test("should display stats cards", async ({ page }: { page: Page }) => {
    await page.goto("/");

    // 检查统计卡片是否存在
    const totalMessages = page.locator("text=Total Messages");
    await expect(totalMessages).toBeVisible();

    const priorityBreakdown = page.locator("text=Priority Breakdown");
    await expect(priorityBreakdown).toBeVisible();

    const typeBreakdown = page.locator("text=Type Breakdown");
    await expect(typeBreakdown).toBeVisible();
  });

  // Skip: Playwright WS mock limitation: does not trigger onopen/onerror events
  test.skip("should handle WebSocket disconnect and reconnect", async ({ page }) => {
    // Mock initial success
    await page.route("**/ws", route => route.fulfill({
      status: 101,
      headers: {
        "Upgrade": "websocket",
        "Connection": "Upgrade",
      },
    }));

    await page.goto("/");

    // Wait for initial connection
    await expect(page.locator("[title='Connected']")).toBeVisible({ timeout: 10000 });

    // Mock WS disconnect
    await page.route("**/ws", route => route.abort());

    // Wait for disconnect state
    await expect(page.locator("text=Disconnected")).toBeVisible({ timeout: 5000 });

    // Mock reconnect success
    await page.route("**/ws", route => route.fulfill({
      status: 101,
      headers: {
        "Upgrade": "websocket",
        "Connection": "Upgrade",
      },
    }));

    // Wait for reconnect
    await expect(page.locator("[title='Connected']")).toBeVisible({ timeout: 10000 });
  });

  // Skip: Playwright WS mock limitation: does not trigger onopen/onerror events
  test.skip("should handle WebSocket error state", async ({ page }) => {
    await page.goto("/");

    // Mock WS error
    await page.route("**/ws", route => route.fulfill({
      status: 500,
    }));

    // Check error state in UI
    await expect(page.locator("text=Connection error")).toBeVisible({ timeout: 5000 });
  });
});
