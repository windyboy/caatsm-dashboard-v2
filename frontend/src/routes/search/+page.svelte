<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import MessageList from "$lib/components/MessageList.svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import { runSearch } from "$lib/api";
  import type { Telegram } from "$lib/types";

  let query = $state("");
  let results = $state<Telegram[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let abortController: AbortController | null = null;
  let debounceTimer: ReturnType<typeof setTimeout> | null = null;
  let initialLoadDone = $state(false);

  async function performSearch() {
    if (abortController) {
      abortController.abort();
    }
    abortController = new AbortController();
    loading = true;
    error = null;

    try {
      const response = await runSearch(query, 100, abortController.signal);
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
    if (debounceTimer) {
      clearTimeout(debounceTimer);
      debounceTimer = null;
    }
    performSearch();
  }

  function reset() {
    query = "";
    if (debounceTimer) {
      clearTimeout(debounceTimer);
      debounceTimer = null;
    }
    performSearch();
  }

  $effect(() => {
    if (!initialLoadDone) return;

    // Access query to ensure effect tracks its changes
    // This ensures the effect re-runs whenever query changes
    query; // eslint-disable-line @typescript-eslint/no-unused-expressions

    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      performSearch();
    }, 400);

    return () => {
      if (debounceTimer) clearTimeout(debounceTimer);
    };
  });

  onMount(() => {
    document.title = "CAATSM Dashboard - Search";
    performSearch().then(() => {
      initialLoadDone = true;
    });
  });

  onDestroy(() => {
    if (debounceTimer) clearTimeout(debounceTimer);
    if (abortController) abortController.abort();
  });
</script>

<main class="page">
  <section class="page__intro">
    <p class="eyebrow">Search</p>
    <h1>Find messages quickly</h1>
    <p class="muted">A focused search view without extra controls or filters.</p>
  </section>

  <form class="search-form" onsubmit={handleSubmit}>
    <label class="search-field">
      <span class="muted">Query</span>
      <input
        name="query"
        placeholder="Type, flight number, route, or text"
        bind:value={query}
        aria-label="Search query"
      />
    </label>
    <div class="search-actions">
      <button class="button" type="submit" disabled={loading}>
        {loading ? "Searching..." : "Search"}
      </button>
      <button class="button button--ghost" type="button" onclick={reset}>
        Clear
      </button>
    </div>
  </form>

  {#if loading}
    <LoadingSkeleton variant="message" />
  {:else if error}
    <div class="card error-card" role="alert" aria-live="polite" aria-atomic="true">
      <p class="eyebrow">Error</p>
      <p class="muted">{error}</p>
    </div>
  {:else}
    <p class="muted">Showing {results.length} of {total} results.</p>
    <MessageList messages={results} />
  {/if}
</main>
