<!--
  @component StatsCard
  Displays statistics for telegram messages in a card format.

  @param {string} title - The title to display on the card
  @param {"total" | "priority" | "type"} type - The type of statistic to display
-->

<script lang="ts">
  import { stats } from "../stores/stats";
  import type { StatsState } from "../stores/stats";

  /** @type {string} */
  export let title: string;
  /** @type {"total" | "priority" | "type"} */
  export let type: "total" | "priority" | "type";

  $: statsValue = $stats as StatsState;
</script>

<div
  class="rounded-lg bg-gradient-to-br from-white/95 to-brand-50/60 p-5 hover:from-brand-50/80 hover:to-accent-50/60 transition-all duration-300 stats-card-glow border-0 hover:shadow-xl hover:scale-105 card-hover-pulse"
>
  <div class="flex items-center gap-3 mb-3">
    <div class="p-2 rounded-lg bg-gradient-to-r from-brand-100 to-accent-100 border border-brand-200/50">
      {#if type === "total"}
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
        </svg>
      {:else if type === "priority"}
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"></path>
        </svg>
      {:else if type === "type"}
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"></path>
        </svg>
      {/if}
    </div>
    <p class="text-xs uppercase tracking-widest text-brand-600 font-bold bg-gradient-to-r from-brand-600 to-accent-600 bg-clip-text text-transparent">{title}</p>
  </div>

  {#if type === "total"}
    <div class="text-2xl font-bold bg-gradient-to-r from-brand-600 via-accent-600 to-success-600 bg-clip-text text-transparent">
      {statsValue.total.toLocaleString()}
    </div>
  {:else if type === "priority"}
    <div class="space-y-2 text-sm text-slate-700 max-h-[150px] overflow-y-auto scrollbar-thin">
      {#each Object.keys(statsValue.byPriority || {}) as priority}
        {@const count = statsValue.byPriority[Number(priority)] || 0}
        <div class="flex items-center justify-between">
          <span>Priority {priority}</span>
          <span class="font-semibold">{count.toLocaleString()}</span>
        </div>
      {/each}
    </div>
  {:else if type === "type"}
    <div class="space-y-2 text-sm text-slate-700 max-h-[200px] overflow-y-auto scrollbar-thin">
      {#each Object.keys(statsValue.byType || {}) as typeName}
        {@const byType = statsValue.byType as Record<string, number>}
        {@const count = byType[typeName] || 0}
        <div class="flex items-center justify-between">
          <span class="uppercase">{typeName}</span>
          <span class="font-semibold">{count.toLocaleString()}</span>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .stats-card-glow {
    background: linear-gradient(
      135deg,
      rgba(249, 250, 251, 0.9) 0%,
      rgba(252, 252, 253, 0.98) 100%
    );
    backdrop-filter: blur(8px);
    box-shadow:
      0 2px 8px rgba(0, 0, 0, 0.03),
      0 1px 2px rgba(0, 0, 0, 0.02),
      0 0 0 0.5px rgba(0, 0, 0, 0.02),
      inset 0 1px 0 rgba(255, 255, 255, 0.95);
    border: 0.5px solid rgba(241, 245, 249, 0.5);
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .stats-card-glow:hover {
    background: linear-gradient(135deg, rgba(241, 245, 249, 0.9) 0%, rgba(255, 255, 255, 1) 100%);
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.06),
      0 2px 4px rgba(0, 0, 0, 0.03),
      0 0 0 0.5px rgba(0, 0, 0, 0.04),
      inset 0 1px 0 rgba(255, 255, 255, 1);
    border-color: rgba(226, 232, 240, 0.7);
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 6px;
  }

  .scrollbar-thin::-webkit-scrollbar-track {
    background: rgba(241, 245, 249, 0.5);
    border-radius: 10px;
  }

  .scrollbar-thin::-webkit-scrollbar-thumb {
    background: rgba(203, 213, 225, 0.6);
    border-radius: 10px;
  }
</style>
