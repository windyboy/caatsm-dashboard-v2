<script lang="ts">
  import Card from "../ui/Card.svelte";
  import MessageCard from "../MessageCard.svelte";
  import type { Telegram } from "$lib/types";

  interface SearchResultsProps {
    results: Telegram[];
    total: number;
    hasSearchCriteria: boolean;
    onReset?: () => void;
  }

  let { results, total, hasSearchCriteria, onReset }: SearchResultsProps = $props();
</script>

<div class="space-y-6 animate-in fade-in duration-200">
  <!-- Results Header -->
  <div class="flex items-center justify-between pb-5 border-b-2 border-slate-200 md:flex-col md:items-start md:gap-4">
    <div>
      <h2 class="text-3xl font-bold text-slate-900 mb-2 md:text-2xl">
        Search Results
      </h2>
      <p class="text-slate-600">
        Found <span class="font-semibold text-slate-900">{total}</span> result{total !== 1 ? 's' : ''}
        {#if total > results.length}
          <span class="text-slate-500"> (showing {results.length})</span>
        {/if}
      </p>
    </div>
    {#if hasSearchCriteria && onReset}
      <button
        type="button"
        onclick={onReset}
        class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-700 bg-white hover:bg-slate-50 border border-slate-300 rounded-lg shadow-sm hover:shadow transition-all md:w-full md:justify-center"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
        Clear filters
      </button>
    {/if}
  </div>

  <!-- Results List -->
  <Card>
    <div
      class="message-list-container max-h-[calc(100vh-480px)] min-h-[500px] overflow-y-auto p-6 bg-gradient-to-b from-slate-50/50 to-white md:max-h-[calc(100vh-420px)] md:min-h-[400px] md:p-4"
    >
      <ul class="list-none p-0 m-0 space-y-4">
        {#each results as item, index (item.message_id ?? `${item.time}-${index}`)}
          <MessageCard message={item} />
        {/each}
      </ul>
    </div>
  </Card>
</div>

<style>
  .message-list-container::-webkit-scrollbar {
    width: 10px;
  }

  .message-list-container::-webkit-scrollbar-track {
    background: #f1f5f9;
    border-radius: 5px;
  }

  .message-list-container::-webkit-scrollbar-thumb {
    background: #cbd5e1;
    border-radius: 5px;
  }

  .message-list-container::-webkit-scrollbar-thumb:hover {
    background: #94a3b8;
  }

  @keyframes fade-in {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-in {
    animation: fade-in 0.3s ease-out;
  }
</style>

