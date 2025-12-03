<script lang="ts">
  import SearchForm from "$lib/components/SearchForm.svelte";
  import SearchResults from "$lib/components/SearchResults.svelte";
  import LoadingSpinner from "$lib/components/ui/LoadingSpinner.svelte";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";
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
  <main class="mx-auto max-w-7xl px-4 sm:px-6 py-6 sm:py-8">
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
{/if}
