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
    
    // Scroll to top when new messages arrive
    unsubscribe = messages.subscribe((msgs: Telegram[]) => {
      if (container && Array.isArray(msgs)) {
        setTimeout(() => {
          container.scrollTo({ top: 0, behavior: "smooth" });
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

<div class="rounded-lg bg-white/90 backdrop-blur-sm border-0 p-8 card-glow">
  <div class="flex items-center justify-between mb-6 pb-4 border-b border-slate-200/60">
    <div class="flex items-center gap-3">
      <div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shadow-md shadow-emerald-500/50"></div>
      <h2 class="text-xl font-bold text-slate-800 tracking-tight">Live Stream</h2>
    </div>
    <span class="text-xs text-emerald-700 bg-emerald-50/80 px-3 py-1 rounded-md font-medium border border-emerald-200/50 shadow-sm">
      Real-time
    </span>
  </div>
  <div
    bind:this={container}
    class="overflow-y-auto scrollbar-thin rounded-lg bg-slate-100/30 p-5 space-y-3 scroll-smooth h-[800px] max-h-[800px] min-h-[800px] border-0 shadow-inner"
  >
    {#if typedMessages.length === 0}
      <p class="text-sm text-slate-400 text-center py-12 font-medium">Waiting for telegrams...</p>
    {:else}
      {#each typedMessages as message (message.message_id)}
        <MessageItem telegram={message} />
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

