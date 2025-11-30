<!--
  @component LiveStream
  Displays a live stream of incoming telegram messages with real-time updates.

  Features:
  - Real-time WebSocket updates
  - Auto-scroll to top for new messages (newest first)
  - Limited message history (50 messages max)
  - Loading state when no messages
-->

<script lang="ts">
  import { messages } from "../stores/messages";
  import { websocket } from "../stores/websocket";
  import MessageItem from "./MessageItem.svelte";
  import type { Telegram } from "../utils/types";
  import { onMount, onDestroy } from "svelte";
  import type { WebSocketStatus } from "../services/websocket";

  let container = $state<HTMLDivElement>();
  let mounted = $state(false);

  // Use local state instead of direct store subscriptions to avoid SSR issues
  let currentStatus = $state<WebSocketStatus>("disconnected");
  let currentReconnectAttempts = $state(0);

  const typedMessages = $derived((Array.isArray($messages) ? $messages : []) as Telegram[]);

  // Generate stable unique key for each message
  // Uses message_id when truthy and not "0", otherwise falls back to time-based key
  const getMessageKey = (message: Telegram, index: number): string => {
    if (message.message_id && message.message_id !== "0") {
      return message.message_id;
    }
    // Fallback: use time-based key for server-scheduled messages without message_id
    // time is stable across re-renders; index ensures uniqueness if time is duplicated
    // Format: scheduled-{time}-{index} or scheduled-{index} if time is missing
    return message.time ? `scheduled-${message.time}-${index}` : `scheduled-${index}`;
  };

  // Connection status for UI - use local state instead of store subscriptions
  const statusColor = $derived.by(() => {
    switch (currentStatus) {
      case "connected":
        return "bg-green-500";
      case "connecting":
        return "bg-yellow-500";
      case "error":
        return "bg-red-500";
      default:
        return "bg-gray-600";
    }
  });

  const statusText = $derived.by(() => {
    switch (currentStatus) {
      case "connected":
        return "Connected";
      case "connecting":
        return "Connecting...";
      case "error":
        return "Connection Error";
      default:
        return "Disconnected";
    }
  });

  // Subscribe to stores only in browser (onMount)
  onMount(() => {
    // Guard against SSR - only run in browser environment
    if (typeof window === "undefined") {
      return;
    }

    mounted = true;

    // Connect WebSocket (idempotent - only connects if not already connected)
    if (!websocket.checkIsConnected()) {
      websocket.connect();
    }

    // Subscribe to status changes
    const statusUnsubscribe = websocket.status.subscribe((status) => {
      currentStatus = status;
    });

    // Subscribe to reconnect attempts
    const reconnectUnsubscribe = websocket.reconnectAttempts.subscribe((attempts) => {
      currentReconnectAttempts = attempts;
    });

    return () => {
      statusUnsubscribe();
      reconnectUnsubscribe();
    };
  });

  // Cleanup WebSocket only on component destroy
  onDestroy(() => {
    websocket.cleanup();
  });

  // Subscribe to messages for auto-scroll (separate from WebSocket lifecycle)
  $effect(() => {
    // Guard against SSR - only run in browser after component is mounted
    if (typeof window === "undefined" || !mounted) {
      return;
    }

    // Scroll to top when new messages arrive (new messages appear at top)
    // Use requestAnimationFrame for smooth DOM updates
    const unsubscribe = messages.subscribe((msgs: Telegram[]) => {
      if (container && Array.isArray(msgs)) {
        requestAnimationFrame(() => {
          if (container) {
            container.scrollTop = 0;
          }
        });
      }
    });

    return () => {
      unsubscribe();
    };
  });
</script>

<div class="live-stream">
  <div class="live-stream-header">
    <div class="live-stream-header-content">
      <div class="status-indicator {statusColor} animate-pulse" title={statusText} aria-hidden="true"></div>
      <svg class="w-5 h-5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-6.938-6.49a8.5 8.5 0 0113.876 0M12 12v4"
        ></path>
      </svg>
      <h2 class="live-stream-title">
        Live Stream
      </h2>
    </div>
    <div class="status-info">
      <span class="status-badge {statusColor}">
        {statusText}
      </span>
      {#if currentReconnectAttempts > 0 && currentStatus !== "connected"}
        <span class="reconnect-attempts">
          (Attempt {currentReconnectAttempts})
        </span>
      {/if}
    </div>
  </div>
  <div bind:this={container} class="messages-container">
    {#if typedMessages.length === 0}
      <div class="empty-state">
        <div class="empty-state-icon"></div>
        <p class="empty-state-text">
          {#if currentStatus === "connecting"}
            Connecting...
          {:else if currentStatus === "error"}
            Connection error. Retrying...
          {:else}
            Waiting for telegrams...
          {/if}
        </p>
        <p class="empty-state-subtitle">Real-time messages will appear here</p>
      </div>
    {:else}
      {#each typedMessages as message, index (getMessageKey(message, index))}
        <div class="message-wrapper">
          <MessageItem telegram={message} />
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .live-stream {
    @apply rounded-lg border-0 p-4 sm:p-8;
    background: var(--card-glow-bg);
    backdrop-filter: var(--card-glow-backdrop);
    box-shadow: var(--card-glow-shadow);
    border: var(--card-glow-border);
    transition: var(--card-glow-transition);
  }

  .live-stream-header {
    @apply flex items-center justify-between mb-4 sm:mb-6 pb-3 sm:pb-4 border-b;
    border-color: theme('colors.brand.200 / 0.4');
  }

  .live-stream-header-content {
    @apply flex items-center gap-3;
  }

  .status-indicator {
    @apply w-2 h-2 rounded-full shadow-lg;
  }

  .live-stream-title {
    @apply text-xl font-bold tracking-tight;
    color: theme('colors.slate.800');
    background: linear-gradient(to right, theme('colors.brand.800'), theme('colors.accent.800'));
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .status-info {
    @apply flex items-center gap-2;
  }

  .status-badge {
    @apply text-xs px-3 py-1 rounded-md font-medium border shadow-md;
  }

  .status-badge.bg-green-500 {
    color: theme('colors.success.700');
    background: linear-gradient(to right, theme('colors.success.50'), theme('colors.brand.50'));
    border-color: theme('colors.success.200 / 0.6');
  }

  .status-badge.bg-yellow-500 {
    color: theme('colors.warning.700');
    background: linear-gradient(to right, theme('colors.warning.50'), theme('colors.warning.50'));
    border-color: theme('colors.warning.200 / 0.6');
  }

  .status-badge.bg-red-500 {
    color: theme('colors.danger.700');
    background: linear-gradient(to right, theme('colors.danger.50'), theme('colors.danger.50'));
    border-color: theme('colors.danger.200 / 0.6');
  }

  .status-badge.bg-gray-600 {
    color: #ffffff !important;
    background-color: #4a5565 !important;
    border-color: rgba(107, 114, 128, 0.8);
  }

  .reconnect-attempts {
    @apply text-xs;
    color: theme('colors.gray.500');
  }

  .messages-container {
    @apply overflow-y-auto scrollbar-thin rounded-lg p-3 sm:p-5 space-y-3 scroll-smooth border-0 shadow-inner;
    background: linear-gradient(to bottom right, theme('colors.slate.50 / 0.4'), theme('colors.brand.50 / 0.3'));
    height: 600px;
    max-height: 600px;
    min-height: 600px;
  }

  @media (min-width: 640px) {
    .messages-container {
      height: 800px;
      max-height: 800px;
      min-height: 800px;
    }
  }

  .empty-state {
    @apply text-center py-12;
  }

  .empty-state-icon {
    @apply w-16 h-16 mx-auto mb-4 rounded-full;
    background: linear-gradient(to right, theme('colors.brand.200'), theme('colors.accent.200'));
  }

  .empty-state-text {
    @apply text-sm font-medium;
    color: theme('colors.slate.500');
  }

  .empty-state-subtitle {
    @apply text-xs mt-1;
    color: theme('colors.slate.400');
  }
</style>
