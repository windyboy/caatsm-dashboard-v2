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
  import { onMount } from "svelte";
  import type { WebSocketStatus } from "../services/websocket";

  let container = $state<HTMLDivElement>();
  let mounted = $state(false);

  // Use local state instead of direct store subscriptions to avoid SSR issues
  let currentStatus = $state<WebSocketStatus>("disconnected");
  let currentReconnectAttempts = $state(0);

  const typedMessages = $derived((Array.isArray($messages) ? $messages : []) as Telegram[]);

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
        return "bg-gray-500";
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
    mounted = true;

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

  // Lifecycle with $effect - only run in browser after mount
  $effect(() => {
    // Guard against SSR - only run in browser after component is mounted
    if (typeof window === "undefined" || !mounted) {
      return;
    }

    websocket.connect();

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
      websocket.cleanup();
    };
  });
</script>

<div   class="rounded-lg bg-white/95 backdrop-blur-md border-0 p-4 sm:p-8 card-glow card-hover-pulse">
  <div class="flex items-center justify-between mb-4 sm:mb-6 pb-3 sm:pb-4 border-b border-brand-200/40">
    <div class="flex items-center gap-3">
      <div
        class="w-2 h-2 rounded-full {statusColor} shadow-lg animate-pulse"
        title={statusText}
      ></div>
      <svg class="w-5 h-5 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-6.938-6.49a8.5 8.5 0 0113.876 0M12 12v4"></path>
      </svg>
      <h2 class="text-xl font-bold text-slate-800 tracking-tight bg-gradient-to-r from-brand-600 to-accent-600 bg-clip-text text-transparent">Live Stream</h2>
    </div>
    <div class="flex items-center gap-2">
      <span
        class="text-xs {statusColor === 'bg-green-500' ? 'text-success-700 bg-gradient-to-r from-success-50 to-brand-50 border-success-200/60' : statusColor === 'bg-yellow-500' ? 'text-yellow-700 bg-gradient-to-r from-yellow-50 to-yellow-50 border-yellow-200/60' : statusColor === 'bg-red-500' ? 'text-red-700 bg-gradient-to-r from-red-50 to-red-50 border-red-200/60' : 'text-gray-700 bg-gradient-to-r from-gray-50 to-gray-50 border-gray-200/60'} px-3 py-1 rounded-md font-medium border shadow-md"
      >
        {statusText}
      </span>
      {#if currentReconnectAttempts > 0 && currentStatus !== "connected"}
        <span class="text-xs text-gray-500">
          (Attempt {currentReconnectAttempts})
        </span>
      {/if}
    </div>
  </div>
  <div
    bind:this={container}
    class="overflow-y-auto scrollbar-thin rounded-lg bg-gradient-to-br from-slate-50/40 to-brand-50/30 p-3 sm:p-5 space-y-3 scroll-smooth h-[600px] sm:h-[800px] max-h-[600px] sm:max-h-[800px] min-h-[600px] sm:min-h-[800px] border-0 shadow-inner"
  >
    {#if typedMessages.length === 0}
      <div class="text-center py-12">
        <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gradient-to-r from-brand-200 to-accent-200"></div>
        <p class="text-sm text-slate-500 font-medium">
          {#if currentStatus === "connecting"}
            Connecting...
          {:else if currentStatus === "error"}
            Connection error. Retrying...
          {:else}
            Waiting for telegrams...
          {/if}
        </p>
        <p class="text-xs text-slate-400 mt-1">Real-time messages will appear here</p>
      </div>
    {:else}
      {#each typedMessages as message, index (message.message_id)}
        <div>
          <MessageItem telegram={message} />
        </div>
      {/each}
    {/if}
  </div>
</div>
