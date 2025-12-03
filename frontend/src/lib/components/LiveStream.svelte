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
  import { messages } from "../stores/data/messages";
  import { websocket } from "../stores/websocket";
  import MessageItem from "./MessageItem.svelte";
  import type { Telegram } from "../utils/types";
  import { onMount, onDestroy } from "svelte";
  import type { WebSocketStatus } from "../services/websocket";
  import VirtualList from "@sveltejs/svelte-virtual-list";

  let container = $state<HTMLDivElement>();
  let mounted = $state(false);

  // Use local state instead of direct store subscriptions to avoid SSR issues
  let currentStatus = $state<WebSocketStatus>("disconnected");
  let currentReconnectAttempts = $state(0);

  const typedMessages = $derived((Array.isArray($messages) ? $messages : []) as Telegram[]);

  // Generate stable unique key for each message
  // Always includes index to ensure uniqueness even if message_id or time are duplicated
  const getMessageKey = (message: Telegram, index: number): string => {
    // Use message_id if available and not "0", but always append index for uniqueness
    if (message.message_id && message.message_id !== "0") {
      return `${message.message_id}-${index}`;
    }
    // Fallback: use time-based key with index to ensure uniqueness
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

    // Subscribe to messages for auto-scroll (new messages appear at top)
    // Use requestAnimationFrame for smooth DOM updates
    const messagesUnsubscribe = messages.subscribe((msgs: Telegram[]) => {
      if (container && Array.isArray(msgs) && msgs.length > 0) {
        requestAnimationFrame(() => {
          if (container) {
            container.scrollTop = 0;
          }
        });
      }
    });

    return () => {
      statusUnsubscribe();
      reconnectUnsubscribe();
      messagesUnsubscribe();
    };
  });

  // Note: WebSocket is a singleton and should persist across component mounts/unmounts
  // We don't cleanup here to avoid disconnecting when the component unmounts
  // The WebSocket connection is managed at the application level
</script>

<div class="live-stream">
  <div class="live-stream-header">
    <div class="live-stream-header-content">
      <div
        class="status-indicator {statusColor} animate-pulse"
        title={statusText}
        aria-hidden="true"
      ></div>
      <svg
        class="w-5 h-5 text-slate-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-6.938-6.49a8.5 8.5 0 0113.876 0M12 12v4"
        ></path>
      </svg>
      <h2 class="live-stream-title">Live Stream</h2>
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
  <div
    bind:this={container}
    class="messages-container"
    aria-live="polite"
    aria-atomic="false"
    aria-relevant="additions"
    role="region"
    aria-label="Live telegram messages"
  >
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
      <VirtualList
        items={typedMessages}
        itemHeight={200}
        height="600px"
        class="virtual-list-container"
        let:item={message}
      >
        <div
          class="message-wrapper"
          role="article"
          aria-label="Telegram message {message.message_id || 'N/A'}"
        >
          <MessageItem telegram={message} />
        </div>
      </VirtualList>
    {/if}
  </div>
</div>

<style>
  .live-stream {
    border-radius: 0.5rem;
    border: var(--card-glow-border);
    padding: 1rem;
    background: var(--card-glow-bg);
    backdrop-filter: var(--card-glow-backdrop);
    box-shadow: var(--card-glow-shadow);
    transition: var(--card-glow-transition);
  }

  .live-stream-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid rgba(var(--brand-200), 0.4);
  }

  .live-stream-header-content {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .status-indicator {
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 9999px;
    box-shadow:
      0 10px 15px -3px rgba(15, 23, 42, 0.15),
      0 4px 6px -4px rgba(15, 23, 42, 0.1);
  }

  .live-stream-title {
    font-size: 1.25rem;
    line-height: 1.75rem;
    font-weight: 700;
    letter-spacing: -0.015em;
    color: rgb(30 41 59);
    background: linear-gradient(to right, rgb(var(--brand-800)), rgb(var(--accent-800)));
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .status-info {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    font-size: 0.75rem;
    line-height: 1rem;
    padding: 0.25rem 0.75rem;
    border-radius: 0.375rem;
    font-weight: 500;
    border: 1px solid transparent;
    box-shadow:
      0 4px 6px -1px rgba(15, 23, 42, 0.1),
      0 2px 4px -2px rgba(15, 23, 42, 0.1);
  }

  .status-badge.bg-green-500 {
    color: #15803d;
    background: linear-gradient(to right, #f0fdf4, #dcfce7);
    border-color: #bbf7d0;
  }

  .status-badge.bg-yellow-500 {
    color: #713f12;
    background: linear-gradient(to right, #fefce8, #fefce8);
    border-color: rgba(254, 240, 138, 0.6);
  }

  .status-badge.bg-red-500 {
    color: #b91c1c;
    background: linear-gradient(to right, #fef2f2, #fef2f2);
    border-color: rgba(254, 202, 202, 0.6);
  }

  .status-badge.bg-gray-600 {
    color: #ffffff !important;
    background-color: #4a5565 !important;
    border-color: rgba(107, 114, 128, 0.8);
  }

  .reconnect-attempts {
    font-size: 0.75rem;
    line-height: 1rem;
    color: #6b7280;
  }

  .messages-container {
    border-radius: 0.5rem;
    padding: 0.75rem;
    border: 0;
    background: linear-gradient(
      to bottom right,
      rgba(248, 250, 252, 0.4),
      rgba(var(--brand-50), 0.3)
    );
    box-shadow: inset 0 2px 4px 0 rgba(15, 23, 42, 0.08);
    height: 600px;
    max-height: 600px;
    min-height: 600px;
  }

  .virtual-list-container {
    width: 100%;
    height: 100%;
  }

  .virtual-list-container :global(.message-wrapper) {
    margin-bottom: 0.75rem;
  }

  .messages-container::-webkit-scrollbar {
    width: 6px;
  }

  .messages-container::-webkit-scrollbar-track {
    background: transparent;
  }

  .messages-container::-webkit-scrollbar-thumb {
    background-color: rgba(148, 163, 184, 0.5);
    border-radius: 9999px;
  }

  .empty-state {
    text-align: center;
    padding: 3rem 0;
  }

  .empty-state-icon {
    width: 4rem;
    height: 4rem;
    margin: 0 auto 1rem;
    border-radius: 9999px;
    background: linear-gradient(to right, rgb(var(--brand-200)), rgb(var(--accent-200)));
  }

  .empty-state-text {
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 500;
    color: #64748b;
  }

  .empty-state-subtitle {
    font-size: 0.75rem;
    line-height: 1rem;
    margin-top: 0.25rem;
    color: #94a3b8;
  }

  @media (min-width: 640px) {
    .live-stream {
      padding: 2rem;
    }

    .live-stream-header {
      margin-bottom: 1.5rem;
      padding-bottom: 1rem;
    }

    .messages-container {
      padding: 1.25rem;
      height: 800px;
      max-height: 800px;
      min-height: 800px;
    }

    .virtual-list-container {
      height: 800px !important;
    }
  }
</style>
