<script lang="ts">
  import SearchForm from "$lib/components/SearchForm.svelte";
  import SearchResults from "$lib/components/SearchResults.svelte";
  import LoadingSpinner from "$lib/components/ui/LoadingSpinner.svelte";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";
  import ThemeToggle from "$lib/components/ThemeToggle.svelte";
  import { search, type SearchParams } from "$lib/services/api";
  import type { Telegram } from "$lib/utils/types";
  import { createLogger } from "$lib/utils/logger";
  import { UI_CONFIG } from "$lib/constants";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";

  import { onMount } from "svelte";

  const logger = createLogger("SearchPage");

  let mounted = $state(false);
  let telegrams = $state<Telegram[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let isSlowRequest = $state(false);
  let loadingTimeout: ReturnType<typeof setTimeout> | null = null;
  let isUpdatingUrl = $state(false); // Track if we're updating URL to prevent effect loop
  let lastSearchParams = $state<string>(""); // Track last search to prevent duplicate searches

  // Helper to get URL params as SearchParams
  function getUrlParams(): SearchParams {
    const params = $page.url.searchParams;
    return {
      query: params.get("query") || undefined,
      type: params.get("type") || undefined,
      priority: params.get("priority") ? parseInt(params.get("priority")!) : undefined,
      start_time: params.get("start_time") || undefined,
      end_time: params.get("end_time") || undefined,
    };
  }

  // Helper to create a search params key for comparison
  function getSearchParamsKey(params: SearchParams): string {
    return JSON.stringify({
      query: params.query || "",
      type: params.type || "",
      priority: params.priority ?? "",
      start_time: params.start_time || "",
      end_time: params.end_time || "",
    });
  }

  // Mount with onMount
  onMount(() => {
    mounted = true;
    
    // Read search parameters from URL on mount
    const urlParams = getUrlParams();
    
    // If URL has params, trigger search
    if (Object.values(urlParams).some((v) => v !== undefined)) {
      const paramsKey = getSearchParamsKey(urlParams);
      lastSearchParams = paramsKey;
      handleSearch(urlParams);
    }
  });

  // React to URL changes (browser back/forward navigation)
  // Only react to URL changes that we didn't initiate ourselves
  $effect(() => {
    if (!mounted || isUpdatingUrl || loading) return;
    
    const urlParams = getUrlParams();
    const paramsKey = getSearchParamsKey(urlParams);
    const hasParams = Object.values(urlParams).some((v) => v !== undefined);
    
    // Only trigger search if params changed (browser navigation)
    // Skip if this is the same search we just performed
    if (hasParams && paramsKey !== lastSearchParams) {
      lastSearchParams = paramsKey;
      handleSearch(urlParams, false); // false = don't update URL (already changed)
    } else if (!hasParams && telegrams.length > 0) {
      // URL cleared - reset results
      lastSearchParams = "";
      handleReset();
    }
  });

  async function handleSearch(params: SearchParams, updateUrl: boolean = true) {
    // Prevent multiple simultaneous searches
    if (loading) {
      logger.warn("Search already in progress, ignoring duplicate request");
      return;
    }

    const paramsKey = getSearchParamsKey(params);
    
    // Skip if this is the same search we just performed (unless it's from browser navigation)
    if (paramsKey === lastSearchParams && !updateUrl) {
      logger.debug("Skipping duplicate search with same parameters");
      return;
    }

    // Update last search params immediately to prevent $effect from retriggering
    lastSearchParams = paramsKey;

    // Update URL first (before search) to enable shareable links
    // Skip URL update if this was triggered by URL change (browser navigation)
    if (updateUrl) {
      isUpdatingUrl = true;
      const searchParams = new URLSearchParams();
      if (params.query) searchParams.set("query", params.query);
      if (params.type) searchParams.set("type", params.type);
      if (params.priority !== undefined) searchParams.set("priority", params.priority.toString());
      if (params.start_time) searchParams.set("start_time", params.start_time);
      if (params.end_time) searchParams.set("end_time", params.end_time);
      
      // Update URL without scrolling and without adding to history if it's the same search
      const newUrl = `/search${searchParams.toString() ? `?${searchParams.toString()}` : ""}`;
      const currentUrl = $page.url.pathname + $page.url.search;
      if (newUrl !== currentUrl) {
        await goto(newUrl, { replaceState: true, noScroll: true });
      }
      // Reset flag after search completes to prevent effect from triggering
      // We'll reset it in the finally block
    }

    loading = true;
    error = null;
    isSlowRequest = false;

    // Clear any existing timeout
    if (loadingTimeout) {
      clearTimeout(loadingTimeout);
      loadingTimeout = null;
    }

    // Show "slow request" warning after timeout
    loadingTimeout = setTimeout(() => {
      isSlowRequest = true;
    }, UI_CONFIG.SLOW_REQUEST_TIMEOUT_MS);

    try {
      logger.debug("Starting search", { params });
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
      const errorMessage = err instanceof Error ? err.message : "Search failed";
      error = errorMessage;
      logger.error("Search failed", err, {
        params: JSON.stringify(params),
        errorMessage,
      });
      // Reset results on error
      telegrams = [];
      total = 0;
    } finally {
      if (loadingTimeout) {
        clearTimeout(loadingTimeout);
        loadingTimeout = null;
      }
      loading = false;
      isSlowRequest = false;
      
      // Reset isUpdatingUrl after search completes (not immediately)
      // This prevents $effect from retriggering right after search completes
      if (updateUrl && isUpdatingUrl) {
        setTimeout(() => {
          isUpdatingUrl = false;
        }, 100);
      }
    }
  }

  function handleReset() {
    telegrams = [];
    total = 0;
    error = null;
    lastSearchParams = "";
    // Clear URL params on reset
    goto("/search", { replaceState: true, noScroll: true });
  }
</script>

{#if mounted}
  <div
    class="min-h-screen text-slate-900 dark:text-slate-100 antialiased bg-gradient-to-br from-blue-50 via-blue-100 to-blue-200 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900"
    style="background-attachment: fixed;"
  >
    <header
      class="bg-white/95 dark:bg-slate-800/95 backdrop-blur-lg shadow-lg border-b border-brand-200/30 dark:border-slate-700/50 sticky top-0 z-50"
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
              class="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight bg-linear-to-r from-brand-600 to-accent-600 dark:from-brand-400 dark:to-accent-400 bg-clip-text text-transparent"
            >
              CAATSM Dashboard
            </h1>
          </div>
          <nav class="flex items-center gap-2">
            <a
              href="/"
              class="px-4 py-2.5 text-sm font-semibold text-slate-700 dark:text-slate-300 hover:text-brand-600 dark:hover:text-brand-400 hover:bg-brand-50/80 dark:hover:bg-slate-700/80 rounded-lg transition-all duration-300 border border-transparent hover:border-brand-200/60 dark:hover:border-slate-600/60 hover:shadow-md hover:scale-105 flex items-center gap-2"
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
              class="px-4 py-2.5 text-sm font-semibold text-slate-700 dark:text-slate-300 hover:text-brand-600 dark:hover:text-brand-400 hover:bg-brand-50/80 dark:hover:bg-slate-700/80 rounded-lg transition-all duration-300 border border-transparent hover:border-brand-200/60 dark:hover:border-slate-600/60 hover:shadow-md hover:scale-105 flex items-center gap-2"
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
            <ThemeToggle />
          </nav>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-7xl px-6 py-8">
      <section class="space-y-8">
        <div>
          <SearchForm
            onsearch={handleSearch}
            onreset={handleReset}
            initialParams={{
              query: $page.url.searchParams.get("query") || "",
              type: $page.url.searchParams.get("type") || "",
              priority: $page.url.searchParams.get("priority") || "",
              start_time: $page.url.searchParams.get("start_time") || "",
              end_time: $page.url.searchParams.get("end_time") || "",
            }}
          />
        </div>

        {#if loading}
          <div class="rounded-lg bg-white/95 dark:bg-slate-800/95 backdrop-blur-md border-0 p-8 card-glow">
            <LoadingSpinner
              size="lg"
              variant="primary"
              message={isSlowRequest
                ? "Still searching... Large result set may take longer."
                : "Searching telegrams..."}
            />
          </div>
        {:else if error}
          <div class="rounded-lg bg-white/95 dark:bg-slate-800/95 backdrop-blur-md border-0 p-8 card-glow">
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
