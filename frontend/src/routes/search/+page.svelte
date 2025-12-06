<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import Input from "$lib/components/ui/Input.svelte";
  import Button from "$lib/components/ui/Button.svelte";
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

<div class="page">
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
    <div class="empty-state">
      <Card>
        <div class="empty-state__content">
          <p class="muted">Enter a search query or time range to search for messages.</p>
        </div>
      </Card>
    </div>
  {:else if results.length === 0}
    <div class="empty-state">
      <Card>
        <div class="empty-state__content">
          <p class="muted">No messages found matching your search criteria.</p>
        </div>
      </Card>
    </div>
  {:else}
    <div class="results-container">
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
    </div>
  {/if}
</div>

<style>
  .page {
    width: 100%;
    max-width: 100%;
    margin: 0;
    padding: 2rem;
    min-height: calc(100vh - 80px);
    background: #fafafa;
  }

  .page__intro {
    margin-bottom: 2.5rem;
    max-width: 1600px;
    margin-left: auto;
    margin-right: auto;
  }

  .page__intro h1 {
    margin: 0.5rem 0 0.75rem;
    font-size: 2.5rem;
    font-weight: 700;
    color: #18181b;
    line-height: 1.2;
  }

  .page__intro .muted {
    font-size: 1.1rem;
    line-height: 1.6;
  }

  .search-form {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr auto;
    gap: 1.25rem;
    align-items: end;
    margin-bottom: 2.5rem;
    padding: 2rem;
    background: white;
    border: 1px solid #e4e4e7;
    border-radius: 12px;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
    max-width: 1600px;
    margin-left: auto;
    margin-right: auto;
  }

  .search-form__time-filters {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .search-actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: nowrap;
    align-items: flex-end;
  }

  .card__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid #e4e4e7;
  }

  .error-card {
    padding: 1.5rem;
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 12px;
    max-width: 1600px;
    margin-left: auto;
    margin-right: auto;
  }

  .error-card .eyebrow {
    color: #dc2626;
    font-weight: 600;
  }

  .message-list-container {
    max-height: calc(100vh - 400px);
    min-height: 500px;
    overflow-y: auto;
    border: 1px solid #e4e4e7;
    border-radius: 12px;
    padding: 1rem;
    background: #fafafa;
  }

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

  .message-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .message {
    padding: 1.5rem;
    background: white;
    border: 1px solid #e4e4e7;
    border-radius: 12px;
    transition: all 0.2s ease;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  }

  .message:hover {
    background: #fafafa;
    border-color: #d4d4d8;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    transform: translateY(-1px);
  }

  .message-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
    flex-wrap: wrap;
  }

  .message-body {
    margin: 1rem 0;
    color: #18181b;
    line-height: 1.7;
    word-break: break-word;
    font-size: 0.95rem;
  }

  .message-foot {
    display: flex;
    align-items: center;
    gap: 1.5rem;
    flex-wrap: wrap;
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e4e4e7;
  }

  .results-container {
    max-width: 1600px;
    margin: 0 auto;
  }

  .empty-state {
    max-width: 1600px;
    margin: 0 auto;
  }

  .empty-state__content {
    padding: 3rem 2rem;
    text-align: center;
  }

  @media (max-width: 1200px) {
    .search-form {
      grid-template-columns: 1fr;
      gap: 1rem;
    }

    .search-form__time-filters {
      grid-template-columns: 1fr 1fr;
    }

    .search-actions {
      justify-content: flex-start;
    }
  }

  @media (max-width: 768px) {
    .page {
      padding: 1rem;
    }

    .page__intro {
      margin-bottom: 1.5rem;
    }

    .page__intro h1 {
      font-size: 1.75rem;
    }

    .search-form {
      padding: 1.25rem;
      margin-bottom: 1.5rem;
    }

    .search-form__time-filters {
      grid-template-columns: 1fr;
    }

    .message-list-container {
      max-height: calc(100vh - 350px);
      min-height: 400px;
      padding: 0.75rem;
    }

    .message {
      padding: 1rem;
    }

    .message-meta,
    .message-foot {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }

    .card__header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }
  }
</style>
