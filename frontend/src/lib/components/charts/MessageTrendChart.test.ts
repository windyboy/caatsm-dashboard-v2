import { describe, it, expect } from 'vitest';
import MessageTrendChart from './MessageTrendChart.svelte';

describe('MessageTrendChart', () => {
  it('should export the component', () => {
    // This is a smoke test to ensure the component can be imported
    // Full rendering tests require client-side environment and Chart.js setup
    // which is better suited for E2E tests
    expect(MessageTrendChart).toBeDefined();
  });
});