<!--
  @component StatsCard
  Displays statistics for telegram messages in a card format.

  @param {string} title - The title to display on the card
  @param {"total" | "priority" | "type"} type - The type of statistic to display
-->

<script lang="ts">
  import { stats } from "../stores/data/stats";
  import type { StatsState } from "../stores/data/stats";

  interface Props {
    title: string;
    type: "total" | "priority" | "type";
  }

  let { title, type }: Props = $props();

  const statsValue = $derived($stats as StatsState);
</script>

<div class="stats-card" role="region" aria-label={title}>
  <div class="stats-header">
    <div class="stats-icon">
      {#if type === "total"}
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          ></path>
        </svg>
      {:else if type === "priority"}
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
          ></path>
        </svg>
      {:else if type === "type"}
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
          ></path>
        </svg>
      {/if}
    </div>
    <p class="stats-title">
      {title}
    </p>
  </div>

  {#if type === "total"}
    <div class="stats-value" aria-live="polite">
      {(statsValue.total ?? 0).toLocaleString()}
    </div>
  {:else if type === "priority"}
    <div class="stats-list" aria-live="polite">
      {#each Object.keys(statsValue.byPriority || {}) as priority (priority)}
        {@const count = statsValue.byPriority[Number(priority)] || 0}
        <div class="stats-list-item">
          <span class="stats-label">Priority {priority}</span>
          <span class="stats-count">{count.toLocaleString()}</span>
        </div>
      {/each}
    </div>
  {:else if type === "type"}
    <div class="stats-list" aria-live="polite">
      {#each Object.keys(statsValue.byType || {}) as typeName (typeName)}
        {@const byType = statsValue.byType as Record<string, number>}
        {@const count = byType[typeName] || 0}
        <div class="stats-list-item">
          <span class="stats-label uppercase">{typeName}</span>
          <span class="stats-count">{count.toLocaleString()}</span>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .stats-card {
    border-radius: 0.5rem;
    padding: 1.25rem;
    transition:
      transform 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    transform: scale(1);
    background: var(--stats-card-bg);
    box-shadow: var(--stats-card-shadow);
    border: var(--stats-card-border);
  }

  .stats-card:hover {
    background: var(--stats-card-hover-bg);
    box-shadow: var(--stats-card-hover-shadow);
    transform: scale(1.05);
  }

  .stats-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  .stats-icon {
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(191, 219, 254, 0.5);
    background: linear-gradient(to right, #dbeafe, #f3e8ff);
  }

  .stats-title {
    font-size: 0.75rem;
    line-height: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 700;
    color: #2563eb;
  }

  .stats-value {
    font-size: 1.5rem;
    line-height: 2rem;
    font-weight: 700;
    color: #2563eb;
  }

  .stats-list {
    font-size: 0.875rem;
    line-height: 1.25rem;
    max-height: 150px;
    overflow-y: auto;
    padding-right: 0.25rem;
    scrollbar-width: thin;
    scrollbar-color: rgba(148, 163, 184, 0.5) transparent;
  }

  .stats-list > * + * {
    margin-top: 0.5rem;
  }

  .stats-list::-webkit-scrollbar {
    width: 6px;
  }

  .stats-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .stats-list::-webkit-scrollbar-thumb {
    background-color: rgba(148, 163, 184, 0.5);
    border-radius: 9999px;
  }

  .stats-list-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .stats-label {
    color: #000000;
    font-weight: 700;
    font-size: 0.9375rem;
    line-height: 1.375rem;
    letter-spacing: 0.01em;
  }

  .stats-count {
    color: #1e3a8a;
    font-weight: 800;
    font-size: 1rem;
    line-height: 1.5rem;
    letter-spacing: 0.02em;
  }

  /* Ensure light mode colors - override any inherited styles */
  :global(:not(.dark)) .stats-label {
    color: #000000;
  }

  :global(:not(.dark)) .stats-count {
    color: #1e3a8a;
  }

  @media (prefers-color-scheme: dark) {
    .stats-title {
      color: #60a5fa;
    }

    .stats-value {
      color: #60a5fa;
    }

    .stats-label {
      color: #ffffff;
      font-weight: 700;
      font-size: 0.9375rem;
      line-height: 1.375rem;
    }

    .stats-count {
      color: #93c5fd;
      font-weight: 800;
      font-size: 1rem;
      line-height: 1.5rem;
    }
  }

  :global(.dark) .stats-title {
    color: #60a5fa;
  }

  :global(.dark) .stats-value {
    color: #60a5fa;
  }

  :global(.dark) .stats-label {
    color: #ffffff;
    font-weight: 700;
    font-size: 0.9375rem;
    line-height: 1.375rem;
  }

  :global(.dark) .stats-count {
    color: #93c5fd;
    font-weight: 800;
    font-size: 1rem;
    line-height: 1.5rem;
  }

  /* Type-specific list height */
  .stats-card:has(.stats-list) .stats-list {
    max-height: 200px;
  }
</style>
