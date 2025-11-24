<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { autocomplete } from "../services/api";
  import { createLogger } from "../utils/logger.ts";

  const logger = createLogger("SearchForm");

  const dispatch = createEventDispatcher();

  let query = "";
  let type = "";
  let priority = "";
  let suggestions: string[] = [];
  let showSuggestions = false;
  let autocompleteTimeout: number | null = null;

  async function handleAutocomplete() {
    if (autocompleteTimeout) {
      clearTimeout(autocompleteTimeout);
    }

    if (query.length < 2) {
      suggestions = [];
      showSuggestions = false;
      return;
    }

    autocompleteTimeout = setTimeout(async () => {
      try {
        const result = await autocomplete(query, 5);
        suggestions = result.suggestions;
        showSuggestions = suggestions.length > 0;
      } catch (error) {
        logger.error("Autocomplete failed", error, {
          query,
          queryLength: query.length,
        });
        suggestions = [];
        showSuggestions = false;
      }
    }, 500);
  }

  function handleSearch() {
    dispatch("search", {
      query,
      type: type || undefined,
      priority: priority ? parseInt(priority) : undefined,
    });
  }

  function selectSuggestion(suggestion: string) {
    query = suggestion;
    showSuggestions = false;
    handleSearch();
  }
</script>

<div class="rounded-lg bg-white/90 backdrop-blur-sm border-0 p-8 card-glow">
  <div class="mb-6 pb-4 border-b border-slate-200/60">
    <h2 class="text-xl font-bold text-slate-800 tracking-tight">Search Telegrams</h2>
  </div>
  <form class="space-y-6" on:submit|preventDefault={handleSearch}>
    <div>
      <label class="block text-sm font-semibold text-slate-700 mb-2.5" for="query">Keywords</label>
      <div class="relative">
        <input
          id="query"
          type="search"
          bind:value={query}
          on:input={handleAutocomplete}
          placeholder="Flight number, message id, content..."
          class="w-full rounded-lg bg-slate-100/50 px-5 py-3 text-slate-900 border-0 placeholder:text-slate-400 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-sky-500/30 focus:shadow-md transition-all duration-200 shadow-sm"
        />
        {#if showSuggestions && suggestions.length > 0}
          <div class="absolute z-10 w-full mt-1 bg-white rounded-lg shadow-lg border border-slate-200 max-h-48 overflow-y-auto">
            {#each suggestions as suggestion (suggestion)}
              <button
                type="button"
                class="w-full text-left px-4 py-2 hover:bg-slate-50 text-sm text-slate-700"
                on:click={() => selectSuggestion(String(suggestion))}
              >
                {suggestion}
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    <div class="grid gap-5 md:grid-cols-3">
      <div>
        <label for="type-select" class="block text-xs uppercase tracking-widest text-slate-600 mb-2.5 font-bold">Type</label>
        <select
          id="type-select"
          bind:value={type}
          class="w-full rounded-lg bg-slate-100/50 px-4 py-3 text-sm text-slate-900 border-0 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-sky-500/30 focus:shadow-md transition-all duration-200 shadow-sm"
        >
          <option value="">Any</option>
          <option value="aftn">AFTN</option>
          <option value="sita">SITA</option>
          <option value="acars">ACARS</option>
          <option value="cpdlc">CPDLC</option>
        </select>
      </div>
      <div>
        <label for="priority-select" class="block text-xs uppercase tracking-widest text-slate-600 mb-2.5 font-bold">Priority</label>
        <select
          id="priority-select"
          bind:value={priority}
          class="w-full rounded-lg bg-slate-100/50 px-4 py-3 text-sm text-slate-900 border-0 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-sky-500/30 focus:shadow-md transition-all duration-200 shadow-sm"
        >
          <option value="">Any</option>
          <option value="1">Urgent</option>
          <option value="2">Operational</option>
          <option value="3">Routine</option>
        </select>
      </div>
      <div>
        <label for="time-range-input" class="block text-xs uppercase tracking-widest text-slate-600 mb-2.5 font-bold">Time Range</label>
        <input
          id="time-range-input"
          type="text"
          placeholder="Last 24h"
          disabled
          title="Coming soon"
          class="w-full rounded-lg bg-slate-100/50 px-4 py-3 text-sm text-slate-900 border-0 placeholder:text-slate-400 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-sky-500/30 focus:shadow-md transition-all duration-200 shadow-sm"
        />
      </div>
    </div>
    <div class="flex justify-end gap-3 pt-4 border-t border-slate-200/60">
      <button
        type="button"
        class="rounded-lg bg-slate-100/80 px-6 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-200/90 transition-all duration-200 shadow-sm hover:shadow border-0"
        on:click={() => {
          query = "";
          type = "";
          priority = "";
          dispatch("reset");
        }}
      >
        Reset
      </button>
      <button
        type="submit"
        class="rounded-lg bg-sky-500 px-6 py-2.5 text-sm font-semibold text-white hover:bg-sky-600 focus:outline-none focus:ring-2 focus:ring-sky-500/40 transition-all duration-200 shadow-md hover:shadow-lg"
      >
        Search
      </button>
    </div>
  </form>
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
</style>

