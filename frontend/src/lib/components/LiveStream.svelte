<!--
  @component LiveStream
  Displays a live stream of incoming telegram messages with real-time updates.

  Features:
  - Real-time WebSocket updates
  - Auto-scroll to bottom for new messages
  - Limited message history (50 messages max)
  - Loading state when no messages
-->

<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { messages } from "../stores/messages";
  import { websocket } from "../stores/websocket";
  import MessageItem from "./MessageItem.svelte";
  import type { Telegram } from "../utils/types";


  let unsubscribe: (() => void) | null = null;

  let container: HTMLDivElement;

  $: typedMessages = (Array.isArray($messages) ? $messages : []) as Telegram[];

  onMount(() => {
    websocket.connect();

    // Scroll to bottom when new messages arrive
    unsubscribe = messages.subscribe((msgs: Telegram[]) => {
      if (container && Array.isArray(msgs)) {
        setTimeout(() => {
          container.scrollTop = container.scrollHeight;
        }, 50);
      }
    });
  });

  onDestroy(() => {
    if (unsubscribe) {
      unsubscribe();
    }
    websocket.cleanup();
  });
</script>

<div   class="rounded-lg bg-white/95 backdrop-blur-md border-0 p-4 sm:p-8 card-glow card-hover-pulse">
  <div class="flex items-center justify-between mb-4 sm:mb-6 pb-3 sm:pb-4 border-b border-brand-200/40">
    <div class="flex items-center gap-3">
      <div
        class="w-2 h-2 rounded-full bg-gradient-to-r from-success-500 to-brand-500 shadow-lg shadow-success-500/60"
      ></div>
      <svg class="w-5 h-5 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-6.938-6.49a8.5 8.5 0 0113.876 0M12 12v4"></path>
      </svg>
      <h2 class="text-xl font-bold text-slate-800 tracking-tight bg-gradient-to-r from-brand-600 to-accent-600 bg-clip-text text-transparent">Live Stream</h2>
    </div>
    <span
      class="text-xs text-success-700 bg-gradient-to-r from-success-50 to-brand-50 px-3 py-1 rounded-md font-medium border border-success-200/60 shadow-md"
    >
      Real-time
    </span>
  </div>
  <div
    bind:this={container}
    class="overflow-y-auto scrollbar-thin rounded-lg bg-gradient-to-br from-slate-50/40 to-brand-50/30 p-3 sm:p-5 space-y-3 scroll-smooth h-[600px] sm:h-[800px] max-h-[600px] sm:max-h-[800px] min-h-[600px] sm:min-h-[800px] border-0 shadow-inner"
  >
    {#if typedMessages.length === 0}
      <div class="text-center py-12">
        <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gradient-to-r from-brand-200 to-accent-200"></div>
        <p class="text-sm text-slate-500 font-medium">Waiting for telegrams...</p>
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

<style>
  .card-glow {
    background: rgba(252, 252, 253, 0.9);
    backdrop-filter: blur(10px);
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.04),
      0 2px 4px rgba(0, 0, 0, 0.02),
      0 0 0 0.5px rgba(0, 0, 0, 0.03),
      inset 0 1px 0 rgba(255, 255, 255, 0.9);
    border: 0.5px solid rgba(226, 232, 240, 0.5);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .card-glow:hover {
    background: rgba(255, 255, 255, 0.95);
    box-shadow:
      0 8px 32px rgba(0, 0, 0, 0.08),
      0 4px 8px rgba(0, 0, 0, 0.04),
      0 0 0 0.5px rgba(0, 0, 0, 0.05),
      inset 0 1px 0 rgba(255, 255, 255, 1);
    border-color: rgba(203, 213, 225, 0.6);
    transform: translateY(-1px);
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  .scrollbar-thin::-webkit-scrollbar-track {
    background: rgba(241, 245, 249, 0.5);
    border-radius: 10px;
  }

  .scrollbar-thin::-webkit-scrollbar-thumb {
    background: rgba(203, 213, 225, 0.6);
    border-radius: 10px;
    border: 2px solid transparent;
    background-clip: padding-box;
    transition: background 0.2s ease;
  }

  .scrollbar-thin::-webkit-scrollbar-thumb:hover {
    background: rgba(148, 163, 184, 0.8);
  }
</style>
