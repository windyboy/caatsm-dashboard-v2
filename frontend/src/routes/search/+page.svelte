<script lang="ts">
  import { onMount } from "svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import SearchForm from "$lib/components/SearchForm.svelte";
  import SearchError from "$lib/components/search/SearchError.svelte";
  import SearchEmptyState from "$lib/components/search/SearchEmptyState.svelte";
  import SearchResults from "$lib/components/search/SearchResults.svelte";
  import { useSearch } from "$lib/composables/useSearch.svelte";

  let query = $state("");
  let startTime = $state("");
  let endTime = $state("");

  const search = useSearch({
    getQuery: () => query,
    getStartTime: () => startTime,
    getEndTime: () => endTime,
  });

  onMount(() => {
    document.title = "CAATSM Dashboard - Search";
    search.handleReset();
  });
</script>

<svelte:head>
  <title>CAATSM Dashboard - Search</title>
</svelte:head>

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
        loading={search.loading}
        onSubmit={search.handleSubmit}
        onReset={search.handleReset}
        onValidationError={search.handleValidationError}
      />
    </div>

    <!-- Error State -->
    {#if search.error}
      <SearchError error={search.error} id="search-form-error" />
    {/if}

    <!-- Loading State -->
    {#if search.loading}
      <div class="animate-in fade-in duration-200">
        <LoadingSkeleton variant="message" />
      </div>
    <!-- Empty States or Results -->
    {:else if search.results.length === 0}
      <SearchEmptyState
        hasSearchCriteria={search.hasSearchCriteria}
        onReset={search.handleReset}
      />
    <!-- Results Section -->
    {:else}
      <SearchResults
        results={search.results}
        total={search.total}
        hasSearchCriteria={search.hasSearchCriteria}
        onReset={search.handleReset}
      />
    {/if}
  </div>
</div>

<style>
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
