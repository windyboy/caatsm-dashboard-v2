<script lang="ts">
  import { onMount } from "svelte";
  import { getConnectionStore } from "$lib/stores/connection.svelte";

  // Lazy load store to avoid circular dependency during module initialization
  let connectionStore = $state<ReturnType<typeof getConnectionStore> | null>(null);

  // Initialize store in onMount instead of $effect to avoid orphan effect error
  onMount(() => {
    if (!connectionStore) {
      connectionStore = getConnectionStore();
    }
  });

  let timeUntilRetry = $state<number | null>(null);
  let intervalId: ReturnType<typeof setInterval> | null = null;

  const status = $derived(connectionStore?.overallStatus ?? "disconnected");
  const nextRetry = $derived(connectionStore?.nextRetryAt ?? null);
  const retryCount = $derived(connectionStore?.retryCount ?? 0);

  $effect(() => {
    if (nextRetry) {
      if (intervalId) clearInterval(intervalId);

      intervalId = setInterval(() => {
        const now = Date.now();
        const retryTime = nextRetry.getTime();
        const diff = Math.max(0, Math.ceil((retryTime - now) / 1000));
        timeUntilRetry = diff > 0 ? diff : null;
      }, 1000);
    } else {
      timeUntilRetry = null;
      if (intervalId) {
        clearInterval(intervalId);
        intervalId = null;
      }
    }

    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  });

  function handleManualRetry() {
    connectionStore?.manualRetry();
    // Trigger a page refresh to retry API calls
    window.location.reload();
  }

  const statusConfig = {
    connected: {
      message: "Connected to server",
      variant: "success",
      show: false,
    },
    degraded: {
      message: "Partially connected, attempting to restore full connection",
      variant: "warning",
      show: true,
    },
    connecting: {
      message: "Connecting to server...",
      variant: "info",
      show: true,
    },
    disconnected: {
      message: "Unable to connect to server",
      variant: "error",
      show: true,
    },
  };

  const config = $derived(statusConfig[status]);
</script>

{#if config.show}
  <div
    class="connection-banner connection-banner--{config.variant}"
    role="alert"
    aria-live="polite"
  >
    <div class="connection-banner__content">
      <span class="connection-banner__icon">
        {#if status === "connecting"}
          <span class="spinner" aria-hidden="true"></span>
        {:else if status === "disconnected"}
          <span aria-hidden="true">⚠️</span>
        {:else}
          <span aria-hidden="true">ℹ️</span>
        {/if}
      </span>
      <div class="connection-banner__text">
        <p class="connection-banner__message">{config.message}</p>
        {#if status === "disconnected" && nextRetry && timeUntilRetry !== null}
          <p class="connection-banner__hint">
            Retrying automatically in {timeUntilRetry} second{timeUntilRetry !== 1 ? "s" : ""}
            {#if retryCount > 0}
              (attempt {retryCount})
            {/if}
          </p>
        {:else if status === "connecting" && retryCount > 0}
          <p class="connection-banner__hint">
            Retrying connection (attempt {retryCount})
          </p>
        {/if}
      </div>
      {#if status === "disconnected"}
        <button class="connection-banner__retry" onclick={handleManualRetry} type="button">
          Retry Now
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .connection-banner {
    position: sticky;
    top: 0;
    z-index: 100;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid;
    backdrop-filter: blur(8px);
    background: rgba(255, 255, 255, 0.95);
  }

  :global(.dark) .connection-banner {
    background: rgba(0, 0, 0, 0.95);
  }

  .connection-banner--success {
    border-color: rgb(34 197 94);
    background: rgba(34, 197, 94, 0.1);
  }

  .connection-banner--warning {
    border-color: rgb(234 179 8);
    background: rgba(234, 179, 8, 0.1);
  }

  .connection-banner--info {
    border-color: rgb(59 130 246);
    background: rgba(59, 130, 246, 0.1);
  }

  .connection-banner--error {
    border-color: rgb(239 68 68);
    background: rgba(239, 68, 68, 0.1);
  }

  .connection-banner__content {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    max-width: 1280px;
    margin: 0 auto;
  }

  .connection-banner__icon {
    display: flex;
    align-items: center;
    flex-shrink: 0;
  }

  .spinner {
    width: 1rem;
    height: 1rem;
    border: 2px solid currentColor;
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .connection-banner__text {
    flex: 1;
    min-width: 0;
  }

  .connection-banner__message {
    font-weight: 500;
    margin: 0;
    font-size: 0.875rem;
  }

  .connection-banner__hint {
    margin: 0.25rem 0 0 0;
    font-size: 0.75rem;
    opacity: 0.8;
  }

  .connection-banner__retry {
    padding: 0.375rem 0.75rem;
    border: 1px solid currentColor;
    border-radius: 0.375rem;
    background: transparent;
    cursor: pointer;
    font-size: 0.875rem;
    transition: opacity 0.2s;
  }

  .connection-banner__retry:hover {
    opacity: 0.8;
  }

  .connection-banner__retry:active {
    opacity: 0.6;
  }
</style>
