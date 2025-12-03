<!--
  @component MetricsCard
  Displays real-time message rate metrics (messages per second).
  
  Features:
  - Current message rate (messages/sec)
  - Peak rate tracking
  - Real-time updates via WebSocket messages
-->

<script lang="ts">
  import { messages } from "$lib/stores/data/messages";
  import type { Telegram } from "$lib/utils/types";
  import { createLogger } from "$lib/utils/logger";
  import { isBrowser } from "$lib/utils/browser";

  const logger = createLogger("MetricsCard");

  let currentRate = $state(0);
  let peakRate = $state(0);
  let messageCount = $state(0);
  let lastUpdateTime = $state<Date | null>(null);
  
  // Track messages in a sliding window (last 10 seconds)
  const RATE_WINDOW_MS = 10000; // 10 seconds
  const messageTimestamps: number[] = [];
  const seenKeys = new Set<string>();
  const recentKeys: string[] = [];
  const MAX_TRACKED_KEYS = 5000;

  function getMessageKey(message: Telegram): string {
    if (message.message_id && message.message_id !== "0") {
      return `id:${message.message_id}`;
    }
    return [
      "hash",
      message.time ?? "",
      message.flight_number ?? "",
      message.source ?? "",
      message.destination ?? "",
      message.type ?? "",
      message.priority ?? "",
      message.content ?? "",
    ].join("|");
  }

  function trackKey(key: string) {
    seenKeys.add(key);
    recentKeys.push(key);
    if (recentKeys.length > MAX_TRACKED_KEYS) {
      const oldest = recentKeys.shift();
      if (oldest) {
        seenKeys.delete(oldest);
      }
    }
  }

  function getMessageTimestamp(message: Telegram): number {
    const parsed = message.time ? new Date(message.time).getTime() : NaN;
    return Number.isNaN(parsed) ? Date.now() : parsed;
  }

  function calculateRate(now: number): void {
    const windowStart = now - RATE_WINDOW_MS;

    // Remove timestamps outside the window (array is not guaranteed sorted, so filter)
    for (let i = messageTimestamps.length - 1; i >= 0; i--) {
      if (messageTimestamps[i] < windowStart) {
        messageTimestamps.splice(i, 1);
      }
    }

    // Calculate rate (messages per second)
    const windowDurationSeconds = RATE_WINDOW_MS / 1000;
    currentRate = messageTimestamps.length / windowDurationSeconds;

    // Update peak rate
    if (currentRate > peakRate) {
      peakRate = currentRate;
    }

    lastUpdateTime = new Date(now);
  }

  function handleNewMessage(message: Telegram): void {
    const key = getMessageKey(message);
    if (seenKeys.has(key)) {
      return; // Skip already processed messages (e.g., from resync batches)
    }

    trackKey(key);
    messageCount++;

    const timestamp = getMessageTimestamp(message);
    const now = Date.now();

    // Ignore old messages for rate calculations to avoid spikes from historical resync
    if (timestamp >= now - RATE_WINDOW_MS) {
      messageTimestamps.push(timestamp);
    }

    calculateRate(now);
  }

  // Subscribe to messages and set up rate calculation using $effect
  $effect(() => {
    if (!isBrowser) {
      return;
    }

    // Subscribe to messages store
    const unsubscribe = messages.subscribe((msgs: Telegram[]) => {
      if (msgs.length === 0) return;

      // Process all messages that have not been seen yet (handles batches)
      for (const msg of msgs) {
        handleNewMessage(msg);
      }
    });

    // Calculate initial rate
    calculateRate(Date.now());

    // Periodic rate recalculation (every second)
    const interval = setInterval(() => {
      calculateRate(Date.now());
    }, 1000);

    // Cleanup subscriptions and interval
    return () => {
      unsubscribe();
      clearInterval(interval);
    };
  });

  function formatRate(rate: number): string {
    if (rate < 0.01) {
      return "0.00";
    }
    return rate.toFixed(2);
  }
</script>

<div class="metrics-card" role="region" aria-label="Message Rate Metrics">
  <div class="metrics-header">
    <div class="metrics-icon">
      <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
        ></path>
      </svg>
    </div>
    <p class="metrics-title">Message Rate</p>
  </div>

  <div class="metrics-content">
    <div class="rate-display">
      <div class="rate-value" aria-live="polite">
        {formatRate(currentRate)}
      </div>
      <div class="rate-unit">messages/sec</div>
    </div>

    <div class="metrics-details">
      <div class="metric-item">
        <span class="metric-label">Peak Rate:</span>
        <span class="metric-value">{formatRate(peakRate)} msg/s</span>
      </div>
      <div class="metric-item">
        <span class="metric-label">Total Messages:</span>
        <span class="metric-value">{messageCount.toLocaleString()}</span>
      </div>
      {#if lastUpdateTime}
        <div class="metric-item">
          <span class="metric-label">Last Update:</span>
          <span class="metric-value">{lastUpdateTime.toLocaleTimeString()}</span>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .metrics-card {
    border-radius: 0.5rem;
    padding: 1.25rem;
    transition:
      transform 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    transform: scale(1);
    background: var(--stats-card-bg);
    box-shadow: var(--stats-card-shadow);
    border: var(--stats-card-border);
  }

  .metrics-card:hover {
    background: var(--stats-card-hover-bg);
    box-shadow: var(--stats-card-hover-shadow);
    transform: scale(1.05);
  }

  .metrics-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  .metrics-icon {
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(191, 219, 254, 0.5);
    background: linear-gradient(to right, #dbeafe, #f3e8ff);
  }

  .metrics-title {
    font-size: 0.75rem;
    line-height: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 700;
    color: #2563eb;
  }

  .metrics-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .rate-display {
    text-align: center;
    padding: 0.5rem 0;
  }

  .rate-value {
    font-size: 2rem;
    line-height: 2.5rem;
    font-weight: 700;
    color: #2563eb;
  }

  .rate-unit {
    font-size: 0.875rem;
    color: #64748b;
    margin-top: 0.25rem;
  }

  .metrics-details {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding-top: 0.75rem;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
  }

  .metric-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.875rem;
  }

  .metric-label {
    color: #64748b;
    font-weight: 500;
  }

  .metric-value {
    color: #334155;
    font-weight: 600;
  }

  @media (prefers-color-scheme: dark) {
    .metrics-title {
      color: #60a5fa;
    }

    .rate-value {
      color: #60a5fa;
    }

    .metric-label {
      color: #94a3b8;
    }

    .metric-value {
      color: #cbd5e1;
    }
  }

  :global(.dark) .metrics-title {
    color: #60a5fa;
  }

  :global(.dark) .rate-value {
    color: #60a5fa;
  }

  :global(.dark) .metric-label {
    color: #94a3b8;
  }

  :global(.dark) .metric-value {
    color: #cbd5e1;
  }
</style>
