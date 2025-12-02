<script lang="ts">
  import SearchForm from "$lib/components/SearchForm.svelte";
  import SearchResults from "$lib/components/SearchResults.svelte";
  import LoadingSpinner from "$lib/components/ui/LoadingSpinner.svelte";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";
  import { search, type SearchParams } from "$lib/services/api";
  import type { Telegram } from "$lib/utils/types";
  import { createLogger } from "$lib/utils/logger";
  import { UI_CONFIG } from "$lib/constants";

  import { onMount } from "svelte";

  const logger = createLogger("SearchPage");

  let mounted = $state(false);
  let telegrams = $state<Telegram[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let isSlowRequest = $state(false);
  let loadingTimeout: ReturnType<typeof setTimeout> | null = null;

  // Mount with onMount
  onMount(() => {
    mounted = true;
  });

  async function handleSearch(params: SearchParams) {
    loading = true;
    error = null;
    isSlowRequest = false;

    // Show "slow request" warning after timeout
    loadingTimeout = setTimeout(() => {
      isSlowRequest = true;
    }, UI_CONFIG.SLOW_REQUEST_TIMEOUT_MS);

    try {
      const result = await search(params);
      // Ensure telegrams is always an array, never null
      telegrams = result.telegrams ?? [];
      total = result.total;
      logger.info("Search succeeded", {
        total: result.total,
        query: params.query || "",
        type: params.type ?? undefined,
        priority: params.priority ?? undefined,
      });
    } catch (err) {
      error = err instanceof Error ? err.message : "Search failed";
      logger.error("Search failed", err, {
        params: JSON.stringify(params),
      });
    } finally {
      if (loadingTimeout) {
        clearTimeout(loadingTimeout);
        loadingTimeout = null;
      }
      loading = false;
      isSlowRequest = false;
    }
  }

  function handleReset() {
    telegrams = [];
    total = 0;
    error = null;
  }
</script>

{#if mounted}
  <div
    class="min-h-screen text-slate-900 antialiased"
    style="background: linear-gradient(135deg, rgb(239, 246, 255) 0%, rgb(219, 234, 254) 25%, rgb(191, 219, 254) 50%, rgb(147, 197, 253) 75%, rgb(96, 165, 250) 100%); background-attachment: fixed;"
  >
    <header
      class="bg-white/95 backdrop-blur-lg shadow-lg border-b border-brand-200/30 sticky top-0 z-50"
    >
      <div class="mx-auto max-w-7xl px-6 py-5">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div
              class="w-10 h-10 rounded-lg bg-linear-to-br from-brand-500 via-accent-500 to-success-500 shadow-lg border border-brand-400/30 flex items-center justify-center"
            >
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                ></path>
              </svg>
            </div>
            <h1
              class="text-2xl font-bold text-slate-900 tracking-tight bg-linear-to-r from-brand-600 to-accent-600 bg-clip-text text-transparent"
            >
              CAATSM Dashboard
            </h1>
          </div>
          <nav class="flex items-center gap-2">
            <a
              href="/"
              class="px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-brand-600 hover:bg-brand-50/80 rounded-lg transition-all duration-300 border border-transparent hover:border-brand-200/60 hover:shadow-md hover:scale-105 flex items-center gap-2"
              aria-label="Dashboard"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                ></path>
              </svg>
              Dashboard
            </a>
            <a
              href="/search"
              class="px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-brand-600 hover:bg-brand-50/80 rounded-lg transition-all duration-300 border border-transparent hover:border-brand-200/60 hover:shadow-md hover:scale-105 flex items-center gap-2"
              aria-label="Search"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                ></path>
              </svg>
              Search
            </a>
          </nav>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-7xl px-6 py-8">
      <section class="space-y-8">
        <div>
          <SearchForm onsearch={handleSearch} onreset={handleReset} />
        </div>

        {#if loading}
          <div class="rounded-lg bg-white/95 backdrop-blur-md border-0 p-8 card-glow">
            <LoadingSpinner
              size="lg"
              variant="primary"
              message={isSlowRequest
                ? "Still searching... Large result set may take longer."
                : "Searching telegrams..."}
            />
          </div>
        {:else if error}
          <div class="rounded-lg bg-white/95 backdrop-blur-md border-0 p-8 card-glow">
            <div class="text-center py-12">
              <p class="text-sm font-semibold text-danger-600 mb-1">Error</p>
              <p class="text-xs text-slate-500">{error}</p>
            </div>
          </div>
        {:else}
          <div>
            <ErrorBoundary context={{ component: "SearchResults" }}>
              <SearchResults {telegrams} {total} />
            </ErrorBoundary>
          </div>
        {/if}
      </section>
    </main>
  </div>
{/if}
