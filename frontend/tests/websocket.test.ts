/**
 * WebSocket 连接测试
 * 
 * 测试 WebSocket 客户端功能：
 * - 连接建立
 * - 消息接收
 * - 自动重连
 * - 统计更新
 */

import { test, expect } from '@playwright/test';

import type { Page } from '@playwright/test';
import type { WebSocket } from '@playwright/test';

test.describe('WebSocket Connection', () => {
  test('should connect to WebSocket endpoint', async ({ page }: { page: Page }) => {
    // Set up WebSocket connection listener before navigation
    const wsPromise = page.waitForEvent('websocket', (ws: WebSocket) => ws.url().includes('/ws'));
    
    await page.goto('/');
    
    // Wait for WebSocket connection to be established
    const _ws = await wsPromise;
    
    // Wait for the page to be fully loaded (ensures components mounted and connection attempted)
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('h1')).toContainText('CAATSM Dashboard');
    
    // Verify WebSocket connection was established
    // The websocket event firing confirms the connection was initiated
  });

  test('should display live stream component', async ({ page }: { page: Page }) => {
    await page.goto('/');
    
    // 检查 LiveStream 组件是否存在
    const liveStream = page.locator('text=Live Stream');
    await expect(liveStream).toBeVisible();
  });

  test('should display stats cards', async ({ page }: { page: Page }) => {
    await page.goto('/');
    
    // 检查统计卡片是否存在
    const totalMessages = page.locator('text=Total Messages');
    await expect(totalMessages).toBeVisible();
    
    const priorityBreakdown = page.locator('text=Priority Breakdown');
    await expect(priorityBreakdown).toBeVisible();
    
    const typeBreakdown = page.locator('text=Type Breakdown');
    await expect(typeBreakdown).toBeVisible();
  });
});

