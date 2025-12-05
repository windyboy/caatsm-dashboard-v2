<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import Input from "$lib/components/ui/Input.svelte";
  import Button from "$lib/components/ui/Button.svelte";
  import Card from "$lib/components/ui/Card.svelte";
  import Badge from "$lib/components/ui/Badge.svelte";
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
    <Card>
      <div class="error-card" role="alert" aria-live="polite" aria-atomic="true">
        <p class="eyebrow">Error</p>
        <p class="muted">{error}</p>
      </div>
    </Card>
  {:else if results.length === 0 && !hasSearchCriteria}
    <Card>
      <p class="muted">Enter a search query or time range to search for messages.</p>
    </Card>
  {:else if results.length === 0}
    <Card>
      <p class="muted">No messages found matching your search criteria.</p>
    </Card>
  {:else}
    <Card>
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
                <Badge variant="soft">{item.type || "Unknown"}</Badge>
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
                <Badge variant="ghost">{formatPriority(item.priority)}</Badge>
              </div>
            </li>
          {/each}
        </ul>
      </div>
    </Card>
  {/if}
</main>

<style>
  .page {
    max-width: 1400px;
    margin: 0 auto;
    padding: 1.5rem;
    min-height: calc(100vh - 80px);
  }

  .page__intro {
    margin-bottom: 2rem;
  }

  .page__intro h1 {
    margin: 0.35rem 0 0.5rem;
    font-size: 2rem;
    font-weight: 700;
    color: #18181b;
  }

  .search-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 2rem;
    padding: 1.5rem;
    background: white;
    border: 1px solid #e4e4e7;
    border-radius: 8px;
  }

  .search-form__time-filters {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 0.75rem;
  }

  .search-actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .card__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .error-card {
    padding: 1rem;
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 8px;
  }

  .error-card .eyebrow {
    color: #dc2626;
  }

  .message-list-container {
    max-height: 600px;
    overflow-y: auto;
    border: 1px solid #e4e4e7;
    border-radius: 8px;
    padding: 0.5rem;
  }

  .message-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .message {
    padding: 1rem;
    background: #fafafa;
    border: 1px solid #e4e4e7;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .message:hover {
    background: #f4f4f5;
    border-color: #d4d4d8;
  }

  .message-meta {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.5rem;
    flex-wrap: wrap;
  }

  .message-body {
    margin: 0.5rem 0;
    color: #18181b;
    line-height: 1.6;
    word-break: break-word;
  }

  .message-foot {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid #e4e4e7;
  }

  @media (max-width: 768px) {
    .page {
      padding: 1rem;
    }

    .page__intro h1 {
      font-size: 1.5rem;
    }

    .search-form {
      padding: 1rem;
    }

    .search-form__time-filters {
      grid-template-columns: 1fr;
    }

    .message-list-container {
      max-height: 400px;
    }

    .message {
      padding: 0.75rem;
    }

    .message-meta,
    .message-foot {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }
  }
</style>
