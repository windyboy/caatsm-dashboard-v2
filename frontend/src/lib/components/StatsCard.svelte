<script lang="ts">
  import { stats } from "../stores/stats";
  import type { StatsState } from "../stores/stats";

  export let title: string;
  export let type: "total" | "priority" | "type";
  
  $: statsValue = $stats as StatsState;
</script>

<div class="rounded-lg bg-gradient-to-br from-slate-50/60 to-white/90 p-5 hover:from-slate-100/70 hover:to-white transition-all duration-200 stats-card-glow border-0">
  <p class="text-xs uppercase tracking-widest text-slate-500 mb-3 font-bold">{title}</p>
  
  {#if type === "total"}
    <div class="text-2xl font-bold text-slate-900">
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
    background: linear-gradient(135deg,
      rgba(249, 250, 251, 0.9) 0%,
      rgba(252, 252, 253, 0.98) 100%);
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
    background: linear-gradient(135deg,
      rgba(241, 245, 249, 0.9) 0%,
      rgba(255, 255, 255, 1) 100%);
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

