<script lang="ts">
  import SearchForm from "$lib/components/SearchForm.svelte";
  import SearchResults from "$lib/components/SearchResults.svelte";
  import { search, type SearchParams } from "$lib/services/api";
  import type { Telegram } from "$lib/utils/types";
  import { createLogger } from "$lib/utils/logger.ts";

  const logger = createLogger("SearchPage");

  let telegrams: Telegram[] = [];
  let total = 0;
  let loading = false;
  let error: string | null = null;

  async function handleSearch(params: SearchParams) {
    loading = true;
    error = null;
    try {
      const result = await search(params);
      telegrams = result.telegrams;
      total = result.total;
    } catch (err) {
      error = err instanceof Error ? err.message : "Search failed";
      logger.error("Search failed", err, {
        params: JSON.stringify(params),
      });
    } finally {
      loading = false;
    }
  }

  function handleReset() {
    telegrams = [];
    total = 0;
    error = null;
  }
</script>

<div class="min-h-screen text-slate-900 antialiased" style="background: linear-gradient(135deg, rgb(238, 242, 247) 0%, rgb(235, 239, 245) 50%, rgb(237, 241, 246) 100%); background-attachment: fixed;">

  <header class="bg-white/95 backdrop-blur-lg shadow-sm border-b border-slate-200/60 sticky top-0 z-50">
    <div class="mx-auto max-w-7xl px-6 py-5">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-sky-500 to-blue-600 shadow-md border border-sky-600/20"></div>
          <h1 class="text-2xl font-bold text-slate-900 tracking-tight">
            CAATSM Dashboard
          </h1>
        </div>
        <nav class="flex items-center gap-2">
          <a href="/" class="px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-sky-600 hover:bg-sky-50/80 rounded-lg transition-all duration-200 border border-transparent hover:border-sky-200/60">
            Dashboard
          </a>
          <a href="/search" class="px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-sky-600 hover:bg-sky-50/80 rounded-lg transition-all duration-200 border border-transparent hover:border-sky-200/60">
            Search
          </a>
        </nav>
      </div>
    </div>
  </header>

  <main class="mx-auto max-w-7xl px-6 py-8">
    <section class="space-y-8">
      <SearchForm on:search={handleSearch} on:reset={handleReset} />

      {#if loading}
        <div class="rounded-lg bg-white/90 backdrop-blur-sm border-0 p-8 card-glow">
          <div class="flex items-center justify-center py-12">
            <div class="skeleton w-full h-8 rounded-md"></div>
          </div>
        </div>
      {:else if error}
        <div class="rounded-lg bg-white/90 backdrop-blur-sm border-0 p-8 card-glow">
          <div class="text-center py-12">
            <p class="text-sm font-semibold text-red-600 mb-1">Error</p>
            <p class="text-xs text-slate-400">{error}</p>
          </div>
        </div>
      {:else}
        <SearchResults {telegrams} {total} />
      {/if}
    </section>
  </main>
</div>

<style>
  .card-glow {
    background: rgba(252, 252, 253, 0.9);
    backdrop-filter: blur(10px);
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.04),
      0 2px 4px rgba(0, 0, 0, 0.02),
      0 0 0 0.5px rgba(0, 0, 0, 0.03),
      inset 0 1px 0 rgba(255, 255, 255, 0.9);
    border: 0.5px solid rgba(226, 232, 240, 0.5);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .skeleton {
    animation: pulse 1.5s ease-in-out infinite;
    background: linear-gradient(90deg,
      rgba(226, 232, 240, 0.8) 0%,
      rgba(241, 245, 249, 0.9) 50%,
      rgba(226, 232, 240, 0.8) 100%);
    background-size: 200% 100%;
  }

  @keyframes pulse {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }
</style>

