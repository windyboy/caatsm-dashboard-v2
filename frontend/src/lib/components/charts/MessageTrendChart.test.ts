import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import MessageTrendChart from './MessageTrendChart.svelte';

describe('MessageTrendChart', () => {
  it('renders chart header', () => {
    render(MessageTrendChart, { target: document.body });
    expect(screen.getByText('Message Trend')).toBeDefined();
    expect(screen.getByText(/Messages over time/)).toBeDefined();
  });
});