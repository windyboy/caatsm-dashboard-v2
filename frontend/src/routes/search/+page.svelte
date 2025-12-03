<script lang="ts">
  import { onMount } from "svelte";
  import MessageList from "$lib/components/MessageList.svelte";
  import { runSearch } from "$lib/api";
  import type { Telegram } from "$lib/types";

  let query = $state("");
  let results = $state<Telegram[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let error = $state<string | null>(null);

  async function performSearch() {
    loading = true;
    error = null;

    try {
      const response = await runSearch(query, 100);
      results = response.telegrams ?? [];
      total = response.total ?? results.length;
    } catch (err) {
      results = [];
      total = 0;
      error = err instanceof Error ? err.message : "Search failed.";
    } finally {
      loading = false;
    }
  }

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    performSearch();
  }

  function reset() {
    query = "";
    performSearch();
  }

  onMount(() => {
    performSearch();
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
    <div class="card">
      <p class="eyebrow">Searching</p>
      <p class="muted">Fetching results...</p>
    </div>
  {:else if error}
    <div class="card error-card">
      <p class="eyebrow">Error</p>
      <p class="muted">{error}</p>
    </div>
  {:else}
    <p class="muted">Showing {results.length} of {total} results.</p>
    <MessageList messages={results} />
  {/if}
</main>
