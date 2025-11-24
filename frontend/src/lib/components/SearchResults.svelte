<script lang="ts">
  import MessageItem from "./MessageItem.svelte";
  import type { Telegram } from "../utils/types";

  export let telegrams: Telegram[] = [];
  export let total: number = 0;
</script>

<div class="rounded-lg bg-white/90 backdrop-blur-sm border-0 p-8 card-glow min-h-[400px]">
  {#if telegrams.length > 0}
    <div class="space-y-5">
      <p class="text-base font-semibold text-slate-700 mb-5">
        Found <span class="text-sky-600 font-bold">{total.toLocaleString()}</span> results
      </p>
      <div class="space-y-4">
        {#each telegrams as telegram (telegram.message_id)}
          <MessageItem telegram={telegram} />
        {/each}
      </div>
    </div>
  {:else}
    <div class="flex flex-col items-center justify-center gap-2 text-slate-400 py-12">
      <p class="text-sm font-medium text-slate-500">No results found</p>
      <p class="text-xs text-slate-400">Try adjusting your search criteria</p>
    </div>
  {/if}
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
</style>

