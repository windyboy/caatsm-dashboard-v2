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
  import { autocomplete } from "../services/api";
  import { createLogger } from "../utils/logger.ts";
  import type { SearchParams } from "../services/api";

  const logger = createLogger("SearchForm");

  interface Props {
    onsearch?: (detail: SearchParams) => void;
    onreset?: () => void;
  }

  let { onsearch, onreset }: Props = $props();

  let query = $state("");
  let type = $state("");
  let priority = $state("");
  let start_time = $state("");
  let end_time = $state("");
  let suggestions = $state<Array<{ value: string; type: string; label: string } | string>>([]);
  let showSuggestions = $state(false);
  let autocompleteTimeout = $state<number | null>(null);

  function toIsoString(value: string): string | undefined {
    if (!value) {
      return undefined;
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return undefined;
    }
    return date.toISOString();
  }

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
        // Handle both old format (string[]) and new format (AutocompleteSuggestion[])
        if (result.suggestions.length > 0 && typeof result.suggestions[0] === "string") {
          // Old format: convert to new format
          suggestions = (result.suggestions as string[]).map((s) => ({
            value: s,
            type: "text",
            label: "",
          }));
        } else {
          suggestions = result.suggestions as Array<{ value: string; type: string; label: string }>;
        }
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
    const startISO = toIsoString(start_time);
    const endISO = toIsoString(end_time);

    onsearch?.({
      query,
      type: type || undefined,
      priority: priority ? parseInt(priority) : undefined,
      start_time: startISO,
      end_time: endISO,
    });
  }

  function selectSuggestion(suggestion: string | { value: string; type: string; label: string }) {
    const value = typeof suggestion === "string" ? suggestion : suggestion.value;
    query = value;
    showSuggestions = false;
    handleSearch();
  }

  function getSuggestionValue(
    suggestion: string | { value: string; type: string; label: string }
  ): string {
    return typeof suggestion === "string" ? suggestion : suggestion.value;
  }

  function getSuggestionLabel(
    suggestion: string | { value: string; type: string; label: string }
  ): string {
    if (typeof suggestion === "string") return "";
    return suggestion.label || suggestion.type || "";
  }
</script>

<div class="rounded-lg bg-white/95 backdrop-blur-md border-0 p-8 card-glow">
  <div class="mb-6 pb-4 border-b border-brand-200/40">
    <div class="flex items-center gap-3">
      <div
        class="p-2 rounded-lg bg-gradient-to-r from-brand-100 to-accent-100 border border-brand-200/50"
      >
        <svg class="w-5 h-5 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          ></path>
        </svg>
      </div>
      <h2
        class="text-xl font-bold text-slate-800 tracking-tight bg-gradient-to-r from-brand-600 to-accent-600 bg-clip-text text-transparent"
      >
        Search Telegrams
      </h2>
    </div>
  </div>
  <form
    class="space-y-6"
    onsubmit={(e) => {
      e.preventDefault();
      handleSearch();
    }}
  >
    <div>
      <label class="block text-sm font-semibold text-slate-700 mb-2.5" for="query">Keywords</label>
      <div class="relative">
        <input
          id="query"
          type="search"
          bind:value={query}
          oninput={handleAutocomplete}
          placeholder="Flight number, message id, content..."
          class="w-full rounded-lg bg-gradient-to-r from-slate-50 to-slate-100/50 px-5 py-3 text-slate-900 border-0 placeholder:text-slate-400 focus:bg-white/95 focus:outline-none focus:ring-2 focus:ring-brand-500/40 focus:shadow-lg focus:shadow-brand-500/20 transition-all duration-300 shadow-sm hover:shadow-md input-focus-glow"
        />
        {#if showSuggestions && suggestions.length > 0}
          <div
            class="absolute z-10 w-full mt-1 bg-white rounded-lg shadow-lg border border-slate-200 max-h-48 overflow-y-auto"
          >
            {#each suggestions as suggestion (getSuggestionValue(suggestion))}
              {@const value = getSuggestionValue(suggestion)}
              {@const label = getSuggestionLabel(suggestion)}
              <button
                type="button"
                class="w-full text-left px-4 py-2 hover:bg-slate-50 text-sm text-slate-700 flex items-center gap-2"
                onclick={() => selectSuggestion(suggestion)}
              >
                {#if label}
                  <span
                    class="text-xs font-semibold text-slate-500 uppercase tracking-wide min-w-[4rem]"
                  >
                    {label}:
                  </span>
                {/if}
                <span class="flex-1 font-medium">{value}</span>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    <div class="grid gap-5 md:grid-cols-2">
      <div class="flex items-center gap-2 mb-2">
        <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
          ></path>
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
          <option value="AFTN">AFTN</option>
          <option value="SITA">SITA</option>
          <option value="ACARS">ACARS</option>
          <option value="CPDLC">CPDLC</option>
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
        onclick={(e) => {
          e.preventDefault();
          query = "";
          type = "";
          priority = "";
          start_time = "";
          end_time = "";
          onreset?.();
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
