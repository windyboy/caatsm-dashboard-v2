<!--
  @component WebSocketMetrics
  Displays WebSocket connection metrics and status.
  
  Features:
  - Connection status
  - Reconnect attempts
  - Connection uptime (if available)
-->

<script lang="ts">
  import { websocket } from "$lib/stores/websocket";
  import type { WebSocketStatus } from "$lib/services/websocket";
  import { createLogger } from "$lib/utils/logger";
  import { isBrowser } from "$lib/utils/browser";

  const logger = createLogger("WebSocketMetrics");

  let status = $state<WebSocketStatus>("disconnected");
  let reconnectAttempts = $state(0);
  let error = $state<string | null>(null);
  let connectionStartTime = $state<Date | null>(null);

  function getStatusColor(status: WebSocketStatus): string {
    switch (status) {
      case "connected":
        return "text-green-800 bg-green-50 border-green-200";
      case "connecting":
      case "reconnecting":
        return "text-yellow-800 bg-yellow-50 border-yellow-200";
      case "error":
        return "text-red-600 bg-red-50 border-red-200";
      default:
        return "text-gray-600 bg-gray-50 border-gray-200";
    }
  }

  function getStatusIcon(status: WebSocketStatus): string {
    switch (status) {
      case "connected":
        return "✓";
      case "connecting":
      case "reconnecting":
        return "⟳";
      case "error":
        return "✗";
      default:
        return "○";
    }
  }

  function getStatusLabel(status: WebSocketStatus): string {
    switch (status) {
      case "connected":
        return "Connected";
      case "connecting":
        return "Connecting";
      case "reconnecting":
        return "Reconnecting";
      case "error":
        return "Error";
      default:
        return "Disconnected";
    }
  }

  function formatUptime(startTime: Date): string {
    const now = new Date();
    const diffMs = now.getTime() - startTime.getTime();
    const diffSeconds = Math.floor(diffMs / 1000);
    const diffMinutes = Math.floor(diffSeconds / 60);
    const diffHours = Math.floor(diffMinutes / 60);

    if (diffHours > 0) {
      return `${diffHours}h ${diffMinutes % 60}m`;
    } else if (diffMinutes > 0) {
      return `${diffMinutes}m ${diffSeconds % 60}s`;
    } else {
      return `${diffSeconds}s`;
    }
  }

  // Subscribe to stores using $effect
  $effect(() => {
    if (!isBrowser) {
      return;
    }

    // Subscribe to status changes
    const statusUnsubscribe = websocket.status.subscribe((newStatus) => {
      status = newStatus;
      
      // Track connection start time
      if (newStatus === "connected" && connectionStartTime === null) {
        connectionStartTime = new Date();
      } else if (newStatus === "disconnected") {
        connectionStartTime = null;
      }
    });

    // Subscribe to reconnect attempts
    const attemptsUnsubscribe = websocket.reconnectAttempts.subscribe((attempts) => {
      reconnectAttempts = attempts;
    });

    // Subscribe to errors
    const errorUnsubscribe = websocket.error.subscribe((err) => {
      error = err;
    });

    // Cleanup subscriptions
    return () => {
      statusUnsubscribe();
      attemptsUnsubscribe();
      errorUnsubscribe();
    };
  });
</script>

<div class="websocket-metrics-card" role="region" aria-label="WebSocket Connection Metrics">
  <div class="metrics-header">
    <div class="metrics-icon">
      <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-6.938-6.49a8.5 8.5 0 0113.876 0M12 12v4"
        ></path>
      </svg>
    </div>
    <p class="metrics-title">WebSocket Status</p>
  </div>

  <div class="metrics-content">
    <!-- Connection Status -->
    <div class="status-display">
      <span class="status-badge {getStatusColor(status)}">
        {getStatusIcon(status)} {getStatusLabel(status)}
      </span>
    </div>

    <!-- Metrics Details -->
    <div class="metrics-details">
      {#if status === "connected" && connectionStartTime}
        <div class="metric-item">
          <span class="metric-label">Uptime:</span>
          <span class="metric-value">{formatUptime(connectionStartTime)}</span>
        </div>
      {/if}

      {#if reconnectAttempts > 0}
        <div class="metric-item">
          <span class="metric-label">Reconnect Attempts:</span>
          <span class="metric-value">{reconnectAttempts}</span>
        </div>
      {/if}

      {#if error}
        <div class="error-message">
          <p class="error-text">{error}</p>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .websocket-metrics-card {
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

  .websocket-metrics-card:hover {
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
    gap: 0.75rem;
  }

  .status-display {
    display: flex;
    justify-content: center;
    margin-bottom: 0.5rem;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 600;
    border: 1px solid;
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

  :global(.dark) .metric-label {
    color: #94a3b8;
  }

  :global(.dark) .metric-value {
    color: #cbd5e1;
  }

  .error-message {
    padding: 0.5rem;
    border-radius: 0.375rem;
    background: rgba(254, 242, 242, 0.5);
    border: 1px solid rgba(254, 202, 202, 0.5);
  }

  .error-text {
    font-size: 0.75rem;
    color: #dc2626;
    text-align: center;
  }
</style>

