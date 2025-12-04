<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import Input from "$lib/components/ui/Input.svelte";
  import Button from "$lib/components/ui/Button.svelte";
  import { runSearch } from "$lib/api";
  import type { Telegram } from "$lib/types";
  // 暂时移除虚拟列表，使用普通 each 循环避免兼容性问题
  // import SvelteVirtualList from "@humanspeak/svelte-virtual-list";

  let query = $state("");
  let startTime = $state("");
  let endTime = $state("");
  let results = $state<Telegram[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let abortController: AbortController | null = null;

  const hasSearchCriteria = $derived.by(() => {
    return query.trim().length > 0 || startTime.length > 0 || endTime.length > 0;
  });

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

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    if (!hasSearchCriteria) {
      error = "Please enter a search query or time range";
      return;
    }
    await performSearch();
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

<main class="page">
  <section class="page__intro">
    <p class="eyebrow">Search</p>
    <h1>Find messages quickly</h1>
    <p class="muted">Search aviation telegrams with full-text search and time range filters.</p>
  </section>

  <form class="search-form" onsubmit={handleSubmit}>
    <Input
      label="Query (optional)"
      name="query"
      placeholder="Type, flight number, route, or text"
      bind:value={query}
      ariaLabel="Search query"
    />
    <div class="search-form__time-filters">
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
    <div class="search-actions">
      <Button type="submit" disabled={loading}>
        {loading ? "Searching..." : "Search"}
      </Button>
      <Button variant="ghost" type="button" onclick={reset}>Clear</Button>
    </div>
  </form>

  {#if loading}
    <LoadingSkeleton variant="message" />
  {:else if error}
    <div class="card error-card" role="alert" aria-live="polite" aria-atomic="true">
      <p class="eyebrow">Error</p>
      <p class="muted">{error}</p>
    </div>
  {:else if results.length === 0 && !hasSearchCriteria}
    <div class="card">
      <p class="muted">Enter a search query or time range to search for messages.</p>
    </div>
  {:else if results.length === 0}
    <div class="card">
      <p class="muted">No messages found matching your search criteria.</p>
    </div>
  {:else}
    <div class="card">
      <div class="card__header">
        <div>
          <p class="eyebrow">Search Results</p>
          <p class="muted">Showing {results.length} of {total} results</p>
        </div>
      </div>
      <div class="message-list-container">
        <ul class="message-list">
          {#each results as item, index (item.message_id ?? `${item.time}-${index}`)}
            <li class="message">
              <div class="message-meta">
                <span class="pill pill--soft">{item.type || "Unknown"}</span>
                <span class="muted">
                  {item.time ? new Date(item.time).toLocaleString() : "No timestamp"}
                </span>
              </div>
              <p class="message-body">{item.content || "No content provided."}</p>
              <div class="message-foot">
                <span class="muted">
                  {item.source || "----"} → {item.destination || "----"}
                </span>
                <span class="muted">{item.flight_number || "Flight N/A"}</span>
                <span class="pill pill--ghost">{formatPriority(item.priority)}</span>
              </div>
            </li>
          {/each}
        </ul>
      </div>
    </div>
  {/if}
</main>

<style>
  .search-form__time-filters {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 0.75rem;
  }
</style>
