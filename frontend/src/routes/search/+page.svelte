<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import Input from "$lib/components/ui/Input.svelte";
  import { Button } from "$lib/components/ui/button/index.js";
  import Card from "$lib/components/ui/Card.svelte";
  import Badge from "$lib/components/ui/Badge.svelte";
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

  function validateTimeRange(): string | null {
    if (startTime && endTime) {
      const start = new Date(startTime);
      const end = new Date(endTime);

      if (start >= end) {
        return "Start time must be before end time";
      }

      const daysDiff = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24);
      if (daysDiff > 90) {
        return "Time range cannot exceed 90 days";
      }
    }
    return null;
  }

  async function performSearch() {
    // Validate time range if both times are provided
    const validationError = validateTimeRange();
    if (validationError) {
      error = validationError;
      results = [];
      total = 0;
      return;
    }

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

    // Validate time range synchronously
    const validationError = validateTimeRange();
    if (validationError) {
      error = validationError;
      return;
    }

    if (!hasSearchCriteria) {
      error = "Please enter a search query or time range";
      return;
    }

    // Perform search asynchronously
    performSearch();
  }

  function reset() {
    query = "";
    startTime = "";
    endTime = "";
    error = null;
    results = [];
    total = 0;
  }

  onMount(() => {
    document.title = "CAATSM Dashboard - Search";
    // Initialize with empty results, don't perform search on mount
    results = [];
    total = 0;
  });

  onDestroy(() => {
    if (abortController) abortController.abort();
  });

  const formatPriority = (priority?: number): string => {
    if (!priority) return "Priority N/A";
    if (priority === 1) return "Priority 1 · Urgent";
    if (priority === 2) return "Priority 2 · Operational";
    return "Priority 3 · Routine";
  };
</script>

<div class="w-full max-w-full m-0 p-8 min-h-[calc(100vh-80px)] bg-[#fafafa] md:p-4">
  <section class="mb-10 max-w-[1600px] mx-auto md:mb-6">
    <p class="eyebrow">Search</p>
    <h1 class="my-2 mt-0.5 mb-0.75 text-4xl font-bold text-[#18181b] leading-tight md:text-3xl">Find messages quickly</h1>
    <p class="muted text-lg leading-relaxed">Search aviation telegrams with full-text search and time range filters.</p>
  </section>

  <form class="search-form grid grid-cols-[2fr_1fr_1fr_auto] gap-5 items-end mb-10 p-8 bg-white border border-[#e4e4e7] rounded-xl shadow-sm max-w-[1600px] mx-auto md:grid-cols-1 md:gap-4 md:p-5 md:mb-6" onsubmit={handleSubmit}>
    <Input
      label="Query (optional)"
      name="query"
      placeholder="Type, flight number, route, or text"
      bind:value={query}
      ariaLabel="Search query"
    />
    <div class="search-form__time-filters grid grid-cols-2 gap-4 md:grid-cols-1">
      <Input
        label="Start Time (optional)"
        name="start_time"
        type="datetime-local"
        bind:value={startTime}
        ariaLabel="Start time"
      />
      <Input
        label="End Time (optional)"
        name="end_time"
        type="datetime-local"
        bind:value={endTime}
        ariaLabel="End time"
      />
    </div>
    <div class="search-actions flex gap-3 flex-nowrap items-end md:justify-start">
      <Button type="submit" disabled={loading}>
        {loading ? "Searching..." : "Search"}
      </Button>
      <Button variant="ghost" type="button" onclick={reset}>Clear</Button>
    </div>
  </form>

  {#if loading}
    <LoadingSkeleton variant="message" />
  {:else if error}
    <Card>
      <div class="p-6 bg-[#fef2f2] border border-[#fecaca] rounded-xl max-w-[1600px] mx-auto" role="alert" aria-live="polite" aria-atomic="true">
        <p class="eyebrow text-[#dc2626] font-semibold">Error</p>
        <p class="muted">{error}</p>
      </div>
    </Card>
  {:else if results.length === 0 && !hasSearchCriteria}
    <div class="max-w-[1600px] mx-auto">
      <Card>
        <div class="py-12 px-8 text-center">
          <p class="muted">Enter a search query or time range to search for messages.</p>
        </div>
      </Card>
    </div>
  {:else if results.length === 0}
    <div class="max-w-[1600px] mx-auto">
      <Card>
        <div class="py-12 px-8 text-center">
          <p class="muted">No messages found matching your search criteria.</p>
        </div>
      </Card>
    </div>
  {:else}
    <div class="max-w-[1600px] mx-auto">
      <Card>
        <div class="card__header flex justify-between items-center gap-4 mb-6 pb-4 border-b-2 border-[#e4e4e7] md:flex-col md:items-start md:gap-2">
          <div>
            <p class="eyebrow">Search Results</p>
            <p class="muted">Showing {results.length} of {total} results</p>
          </div>
        </div>
        <div class="message-list-container max-h-[calc(100vh-400px)] min-h-[500px] overflow-y-auto border border-[#e4e4e7] rounded-xl p-4 bg-[#fafafa] md:max-h-[calc(100vh-350px)] md:min-h-[400px] md:p-3">
          <ul class="list-none p-0 m-0 flex flex-col gap-4">
            {#each results as item, index (item.message_id ?? `${item.time}-${index}`)}
              <li class="message p-6 bg-white border border-[#e4e4e7] rounded-xl transition-all shadow-sm hover:bg-[#fafafa] hover:border-[#d4d4d8] hover:shadow-md hover:-translate-y-px md:p-4">
                <div class="message-meta flex items-center gap-4 mb-4 flex-wrap md:flex-col md:items-start md:gap-2">
                  <Badge variant="soft">{item.type || "Unknown"}</Badge>
                  <span class="muted">
                    {item.time ? new Date(item.time).toLocaleString() : "No timestamp"}
                  </span>
                </div>
                <p class="message-body my-4 text-[#18181b] leading-relaxed break-words text-[0.95rem]">{item.content || "No content provided."}</p>
                <div class="message-foot flex items-center gap-6 flex-wrap mt-4 pt-4 border-t border-[#e4e4e7] md:flex-col md:items-start md:gap-2">
                  <span class="muted">
                    {item.source || "----"} → {item.destination || "----"}
                  </span>
                  <span class="muted">{item.flight_number || "Flight N/A"}</span>
                  <Badge variant="ghost">{formatPriority(item.priority)}</Badge>
                </div>
              </li>
            {/each}
          </ul>
        </div>
      </Card>
    </div>
  {/if}
</div>

<style>
  /* Custom scrollbar styles - keep as scoped styles */
  .message-list-container::-webkit-scrollbar {
    width: 8px;
  }

  .message-list-container::-webkit-scrollbar-track {
    background: #f4f4f5;
    border-radius: 4px;
  }

  .message-list-container::-webkit-scrollbar-thumb {
    background: #d4d4d8;
    border-radius: 4px;
  }

  .message-list-container::-webkit-scrollbar-thumb:hover {
    background: #a1a1aa;
  }
</style>
