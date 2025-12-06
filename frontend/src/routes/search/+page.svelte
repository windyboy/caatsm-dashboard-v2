<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import Card from "$lib/components/ui/Card.svelte";
  import SearchForm from "$lib/components/SearchForm.svelte";
  import MessageCard from "$lib/components/MessageCard.svelte";
  import { runSearch } from "$lib/api";
  import type { Telegram } from "$lib/types";

  let query = $state("");
  let startTime = $state("");
  let endTime = $state("");
  let results = $state<Telegram[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let abortController: AbortController | null = null;

  const hasSearchCriteria = $derived(
    query.trim().length > 0 || startTime.length > 0 || endTime.length > 0
  );

  async function performSearch() {
    if (abortController) {
      abortController.abort();
    }
    abortController = new AbortController();
    loading = true;
    error = null;

    try {
      const response = await runSearch(
        query,
        1000,
        abortController.signal,
        startTime || undefined,
        endTime || undefined
      );
      if (abortController.signal.aborted) return;
      results = response.telegrams ?? [];
      total = response.total ?? results.length;
    } catch (err) {
      if (err instanceof Error && err.name === "AbortError") return;
      results = [];
      total = 0;
      error = err instanceof Error ? err.message : "Search failed.";
    } finally {
      if (!abortController.signal.aborted) {
        loading = false;
      }
    }
  }

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();

    if (!hasSearchCriteria) {
      error = "Please enter a search query or time range";
      return;
    }

    performSearch();
  }

  function handleValidationError(validationError: string) {
    error = validationError;
    results = [];
    total = 0;
  }

  function handleReset() {
    error = null;
    results = [];
    total = 0;
  }

  onMount(() => {
    document.title = "CAATSM Dashboard - Search";
    results = [];
    total = 0;
  });

  onDestroy(() => {
    if (abortController) abortController.abort();
  });
</script>

<div class="w-full min-h-[calc(100vh-80px)] bg-gradient-to-br from-slate-50 via-white to-blue-50/30">
  <div class="max-w-[1600px] mx-auto px-6 py-10 md:px-4 md:py-6">
    <!-- Header Section with improved design -->
    <header class="mb-8 md:mb-6">
      <div class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-blue-100 text-blue-700 rounded-full text-xs font-medium mb-3">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        Search
      </div>
      <h1 class="text-4xl font-bold text-slate-900 leading-tight mb-3 md:text-3xl">
        Find messages quickly
      </h1>
      <p class="text-base text-slate-600 leading-relaxed max-w-2xl">
        Search aviation telegrams with full-text search and time range filters. Get instant results from your message database.
      </p>
    </header>

    <!-- Search Form Section -->
    <div class="mb-8">
      <SearchForm
        bind:query
        bind:startTime
        bind:endTime
        {loading}
        onSubmit={handleSubmit}
        onReset={handleReset}
        onValidationError={handleValidationError}
      />
    </div>

    <!-- Error State -->
    {#if error}
      <div class="mb-6 animate-in fade-in duration-200">
        <div
          class="p-5 bg-red-50 border-l-4 border-red-500 rounded-lg shadow-sm"
          role="alert"
          aria-live="polite"
          aria-atomic="true"
        >
          <div class="flex items-start gap-3">
            <div class="flex-shrink-0">
              <svg class="w-5 h-5 text-red-600 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div class="flex-1">
              <h3 class="text-sm font-semibold text-red-900 mb-1">Error</h3>
              <p class="text-sm text-red-800">{error}</p>
            </div>
          </div>
        </div>
      </div>
    {/if}

    <!-- Loading State -->
    {#if loading}
      <div class="animate-in fade-in duration-200">
        <LoadingSkeleton variant="message" />
      </div>
    <!-- Empty States -->
    {:else if results.length === 0 && !hasSearchCriteria}
      <Card>
        <div class="py-10 px-6 text-center">
          <div class="max-w-md mx-auto">
            <div class="w-14 h-14 mx-auto mb-4 bg-blue-50 rounded-xl flex items-center justify-center border border-blue-100">
              <svg class="w-7 h-7 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <h2 class="text-base font-semibold text-slate-700 mb-2">
              Ready to search
            </h2>
            <p class="text-slate-500 mb-6 text-xs leading-relaxed">
              Enter a search query or use the time filters above to find messages in your database.
            </p>
            <div class="grid grid-cols-3 gap-2 max-w-xs mx-auto">
              <div class="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                <div class="font-medium text-slate-700 mb-0.5 text-[11px]">Full-text</div>
                <div class="text-slate-500 text-[10px] leading-tight">Search content</div>
              </div>
              <div class="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                <div class="font-medium text-slate-700 mb-0.5 text-[11px]">Time range</div>
                <div class="text-slate-500 text-[10px] leading-tight">Filter by date</div>
              </div>
              <div class="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                <div class="font-medium text-slate-700 mb-0.5 text-[11px]">90 days</div>
                <div class="text-slate-500 text-[10px] leading-tight">Max range</div>
              </div>
            </div>
          </div>
        </div>
      </Card>
    {:else if results.length === 0}
      <Card>
        <div class="py-10 px-6 text-center">
          <div class="max-w-md mx-auto">
            <div class="w-14 h-14 mx-auto mb-4 bg-amber-50 rounded-xl flex items-center justify-center border border-amber-100">
              <svg class="w-7 h-7 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h2 class="text-base font-semibold text-slate-700 mb-2">
              No results found
            </h2>
            <p class="text-slate-500 mb-6 text-xs leading-relaxed">
              No messages match your search criteria. Try adjusting your filters or search terms.
            </p>
            <button
              type="button"
              onclick={handleReset}
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm hover:shadow-md transition-all"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
              Clear search
            </button>
          </div>
        </div>
      </Card>
    <!-- Results Section -->
    {:else}
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
          {#if hasSearchCriteria}
            <button
              type="button"
              onclick={handleReset}
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
    {/if}
  </div>
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
