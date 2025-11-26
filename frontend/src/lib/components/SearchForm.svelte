<!--
  @component SearchForm
  Search form for filtering telegram messages with autocomplete and filters.

  Features:
  - Keyword search with autocomplete
  - Type, priority, and time range filters
  - Debounced autocomplete (500ms)
  - Form validation and reset functionality

  @event {SearchParams} search - Dispatched when search is performed
  @event reset - Dispatched when form is reset
-->

<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { autocomplete } from "../services/api";
  import { createLogger } from "../utils/logger.ts";


  const logger = createLogger("SearchForm");

  const dispatch = createEventDispatcher();

  let query = "";
  let type = "";
  let priority = "";
  let start_time = "";
  let end_time = "";
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
      start_time: start_time || undefined,
      end_time: end_time || undefined,
    });
  }

  function selectSuggestion(suggestion: string) {
    query = suggestion;
    showSuggestions = false;
    handleSearch();
  }
</script>

<div   class="rounded-lg bg-white/95 backdrop-blur-md border-0 p-8 card-glow">
  <div class="mb-6 pb-4 border-b border-brand-200/40">
    <div class="flex items-center gap-3">
      <div class="p-2 rounded-lg bg-gradient-to-r from-brand-100 to-accent-100 border border-brand-200/50">
        <svg class="w-5 h-5 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
        </svg>
      </div>
      <h2 class="text-xl font-bold text-slate-800 tracking-tight bg-gradient-to-r from-brand-600 to-accent-600 bg-clip-text text-transparent">Search Telegrams</h2>
    </div>
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
          class="w-full rounded-lg bg-gradient-to-r from-slate-50 to-slate-100/50 px-5 py-3 text-slate-900 border-0 placeholder:text-slate-400 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:shadow-lg focus:shadow-brand-500/20 transition-all duration-300 shadow-sm hover:shadow-md input-focus-glow"
        />
        {#if showSuggestions && suggestions.length > 0}
          <div
            class="absolute z-10 w-full mt-1 bg-white rounded-lg shadow-lg border border-slate-200 max-h-48 overflow-y-auto"
          >
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
    <div class="grid gap-5 md:grid-cols-2">
      <div class="flex items-center gap-2 mb-2">
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"></path>
        </svg>
        <span class="text-sm font-semibold text-brand-700">Filters</span>
      </div>
      <div></div>
      <div>
        <label
          for="type-select"
          class="block text-xs uppercase tracking-widest text-slate-600 mb-2.5 font-bold"
          >Type</label
        >
        <select
          id="type-select"
          bind:value={type}
          class="w-full rounded-lg bg-gradient-to-r from-slate-50 to-slate-100/50 px-4 py-3 text-sm text-slate-900 border-0 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:shadow-lg focus:shadow-brand-500/20 transition-all duration-300 shadow-sm hover:shadow-md input-focus-glow"
        >
          <option value="">Any</option>
          <option value="aftn">AFTN</option>
          <option value="sita">SITA</option>
          <option value="acars">ACARS</option>
          <option value="cpdlc">CPDLC</option>
        </select>
      </div>
      <div>
        <label
          for="priority-select"
          class="block text-xs uppercase tracking-widest text-slate-600 mb-2.5 font-bold"
          >Priority</label
        >
        <select
          id="priority-select"
          bind:value={priority}
          class="w-full rounded-lg bg-gradient-to-r from-slate-50 to-slate-100/50 px-4 py-3 text-sm text-slate-900 border-0 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:shadow-lg focus:shadow-brand-500/20 transition-all duration-300 shadow-sm hover:shadow-md input-focus-glow"
        >
          <option value="">Any</option>
          <option value="1">Urgent</option>
          <option value="2">Operational</option>
          <option value="3">Routine</option>
        </select>
      </div>
      <div>
        <label
          for="start-time-input"
          class="block text-xs uppercase tracking-widest text-slate-600 mb-2.5 font-bold"
          >Start Time</label
        >
        <input
          id="start-time-input"
          type="datetime-local"
          bind:value={start_time}
          class="w-full rounded-lg bg-gradient-to-r from-slate-50 to-slate-100/50 px-4 py-3 text-sm text-slate-900 border-0 placeholder:text-slate-400 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:shadow-lg focus:shadow-brand-500/20 transition-all duration-300 shadow-sm hover:shadow-md input-focus-glow"
        />
      </div>
      <div>
        <label
          for="end-time-input"
          class="block text-xs uppercase tracking-widest text-slate-600 mb-2.5 font-bold"
          >End Time</label
        >
        <input
          id="end-time-input"
          type="datetime-local"
          bind:value={end_time}
          class="w-full rounded-lg bg-gradient-to-r from-slate-50 to-slate-100/50 px-4 py-3 text-sm text-slate-900 border-0 placeholder:text-slate-400 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:shadow-lg focus:shadow-brand-500/20 transition-all duration-300 shadow-sm hover:shadow-md input-focus-glow"
        />
      </div>
    </div>
    <div class="flex justify-end gap-3 pt-4 border-t border-slate-200/60">
        <button
          type="button"
          class="rounded-lg bg-gradient-to-r from-slate-100/80 to-slate-200/60 px-6 py-2.5 text-sm font-semibold text-slate-700 hover:from-slate-200/90 hover:to-slate-300/70 transition-all duration-300 shadow-sm hover:shadow-md hover:scale-105 border-0 ripple btn-hover-lift"
          on:click|preventDefault={() => {
            query = "";
            type = "";
            priority = "";
            start_time = "";
            end_time = "";
            dispatch("reset");
          }}
        >
         Reset
       </button>
       <button
         type="submit"
         class="rounded-lg bg-gradient-to-r from-brand-500 to-accent-500 px-6 py-2.5 text-sm font-semibold text-white hover:from-brand-600 hover:to-accent-600 focus:outline-none focus:ring-2 focus:ring-brand-500/40 transition-all duration-300 shadow-lg hover:shadow-xl hover:scale-105 ripple btn-hover-lift"
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
