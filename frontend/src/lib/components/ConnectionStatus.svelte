<script lang="ts">
  import { websocket } from '../stores/websocket';
  import { onMount } from 'svelte';

  let status: string = 'disconnected';
  let error: string | null = null;
  let reconnectAttempts: number = 0;
  let isVisible = false;

  // Subscribe to stores
  const unsubscribeStatus = websocket.status.subscribe(value => {
    status = value;
    // Always show for debugging and testing
    isVisible = true;
  });

  const unsubscribeError = websocket.error.subscribe(value => {
    error = value;
  });

  const unsubscribeAttempts = websocket.reconnectAttempts.subscribe(value => {
    reconnectAttempts = value;
  });

  onMount(() => {
    // Auto-hide after successful connection
    const unsubscribe = websocket.status.subscribe(currentStatus => {
      if (currentStatus === 'connected' && status !== 'connected') {
        setTimeout(() => {
          isVisible = false;
        }, 2000); // Hide after 2 seconds
      }
    });

    return unsubscribe;
  });

  function getStatusIcon() {
    switch (status) {
      case 'connecting':
        return '🔄';
      case 'connected':
        return '🟢';
      case 'reconnecting':
        return '🔄';
      case 'error':
        return '🔴';
      case 'disconnected':
        return '⚪';
      default:
        return '⚪';
    }
  }

  function getStatusText() {
    switch (status) {
      case 'connecting':
        return 'Connecting...';
      case 'connected':
        return 'Connected';
      case 'reconnecting':
        return `Reconnecting... (${reconnectAttempts})`;
      case 'error':
        return 'Connection Error';
      case 'disconnected':
        return 'Disconnected';
      default:
        return 'Unknown';
    }
  }

  function getStatusClass() {
    switch (status) {
      case 'connecting':
      case 'reconnecting':
        return 'status-connecting';
      case 'connected':
        return 'status-connected';
      case 'error':
        return 'status-error';
      case 'disconnected':
        return 'status-disconnected';
      default:
        return 'status-unknown';
    }
  }
</script>

{#if isVisible || error}
  <div class="connection-status {getStatusClass()}" title={getStatusText()} role="status" aria-live="polite">
    <span class="status-icon">{getStatusIcon()}</span>
    <span class="status-text">{getStatusText()}</span>
    {#if error}
      <div class="error-message">{error}</div>
    {/if}
  </div>
{/if}

<style>
  .connection-status {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 1000;
    padding: 8px 16px;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    backdrop-filter: blur(8px);
    transition: all 0.3s ease;
    max-width: 300px;
  }

  .status-icon {
    font-size: 16px;
    flex-shrink: 0;
  }

  .status-text {
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .error-message {
    margin-top: 4px;
    font-size: 12px;
    font-weight: 400;
    color: inherit;
    opacity: 0.9;
    line-height: 1.4;
  }

  .status-connecting {
    background: rgba(59, 130, 246, 0.9);
    color: white;
    border: 1px solid rgba(59, 130, 246, 0.3);
  }

  .status-connected {
    background: rgba(34, 197, 94, 0.9);
    color: white;
    border: 1px solid rgba(34, 197, 94, 0.3);
  }

  .status-error {
    background: rgba(239, 68, 68, 0.9);
    color: white;
    border: 1px solid rgba(239, 68, 68, 0.3);
  }

  .status-disconnected {
    background: rgba(75, 85, 99, 0.9);
    color: white;
    border: 1px solid rgba(75, 85, 99, 0.3);
  }

  .status-unknown {
    background: rgba(107, 114, 128, 0.9);
    color: white;
    border: 1px solid rgba(107, 114, 128, 0.3);
  }

  /* Animation for connecting state */
  .status-connecting .status-icon {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  /* Responsive design */
  @media (max-width: 640px) {
    .connection-status {
      top: 10px;
      right: 10px;
      left: 10px;
      max-width: none;
      padding: 6px 12px;
      font-size: 13px;
    }

    .status-icon {
      font-size: 14px;
    }
  }
</style>