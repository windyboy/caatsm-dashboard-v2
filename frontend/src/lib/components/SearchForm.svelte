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
  import { createLogger } from "../utils/logger";
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

<div class="search-form">
  <div class="search-header">
    <div class="search-header-content">
      <div class="search-icon">
        <svg class="w-5 h-5 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          ></path>
        </svg>
      </div>
      <h2 class="search-title">
        Search Telegrams
      </h2>
    </div>
  </div>
  <form
    class="search-form-fields"
    onsubmit={(e) => {
      e.preventDefault();
      handleSearch();
    }}
  >
    <div class="search-field">
      <label class="search-label" for="query">Keywords</label>
      <div class="search-input-container">
        <input
          id="query"
          type="search"
          bind:value={query}
          oninput={handleAutocomplete}
          placeholder="Flight number, message id, content..."
          class="search-input"
        />
        {#if showSuggestions && suggestions.length > 0}
          <div class="suggestions-dropdown">
            {#each suggestions as suggestion (getSuggestionValue(suggestion))}
              {@const value = getSuggestionValue(suggestion)}
              {@const label = getSuggestionLabel(suggestion)}
              <button
                type="button"
                class="suggestion-item"
                onclick={() => selectSuggestion(suggestion)}
              >
                {#if label}
                  <span class="suggestion-label">
                    {label}:
                  </span>
                {/if}
                <span class="suggestion-value">{value}</span>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    <div class="filters-grid">
      <div class="filters-header">
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
      <div class="filter-field">
        <label for="type-select" class="filter-label">Type</label>
        <select id="type-select" bind:value={type} class="filter-select">
          <option value="">Any</option>
          <option value="AFTN">AFTN</option>
          <option value="SITA">SITA</option>
          <option value="ACARS">ACARS</option>
          <option value="CPDLC">CPDLC</option>
        </select>
      </div>
      <div class="filter-field">
        <label for="priority-select" class="filter-label">Priority</label>
        <select id="priority-select" bind:value={priority} class="filter-select">
          <option value="">Any</option>
          <option value="1">Urgent</option>
          <option value="2">Operational</option>
          <option value="3">Routine</option>
        </select>
      </div>
      <div class="filter-field">
        <label for="start-time-input" class="filter-label">Start Time</label>
        <input
          id="start-time-input"
          type="datetime-local"
          bind:value={start_time}
          class="filter-input"
        />
      </div>
      <div class="filter-field">
        <label for="end-time-input" class="filter-label">End Time</label>
        <input
          id="end-time-input"
          type="datetime-local"
          bind:value={end_time}
          class="filter-input"
        />
      </div>
    </div>
    <div class="form-actions">
      <button
        type="button"
        class="btn-reset"
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
      <button type="submit" class="btn-search">
        Search
      </button>
    </div>
  </form>
</div>

<style>
  .search-form {
    @apply rounded-lg border-0 p-8;
    background: var(--card-glow-bg);
    backdrop-filter: var(--card-glow-backdrop);
    box-shadow: var(--card-glow-shadow);
    border: var(--card-glow-border);
    transition: var(--card-glow-transition);
  }

  .search-header {
    @apply mb-6 pb-4 border-b;
    border-color: theme('colors.brand.200 / 0.4');
  }

  .search-header-content {
    @apply flex items-center gap-3;
  }

  .search-icon {
    @apply p-2 rounded-lg border;
    background: linear-gradient(to right, theme('colors.brand.100'), theme('colors.accent.100'));
    border-color: theme('colors.brand.200 / 0.5');
  }

  .search-title {
    @apply text-xl font-bold tracking-tight;
    color: theme('colors.slate.800');
    background: linear-gradient(to right, theme('colors.brand.600'), theme('colors.accent.600'));
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .search-form-fields {
    @apply space-y-6;
  }

  .search-field {
    /* No specific styles needed */
  }

  .search-label {
    @apply block text-sm font-semibold mb-2.5;
    color: theme('colors.slate.700');
  }

  .search-input-container {
    @apply relative;
  }

  .search-input {
    @apply w-full rounded-lg px-5 py-3 text-slate-900 border-0 placeholder:text-slate-400 transition-all duration-300 shadow-sm hover:shadow-md;
    background: linear-gradient(to right, theme('colors.slate.50'), theme('colors.slate.100 / 0.5'));
  }

  .search-input:focus {
    @apply bg-white/95 outline-none shadow-lg;
    box-shadow: var(--input-focus-shadow);
  }

  .suggestions-dropdown {
    @apply absolute z-10 w-full mt-1 bg-white rounded-lg shadow-lg border max-h-48 overflow-y-auto;
    border-color: theme('colors.slate.200');
  }

  .suggestion-item {
    @apply w-full text-left px-4 py-2 text-sm flex items-center gap-2;
    color: theme('colors.slate.700');
  }

  .suggestion-item:hover {
    background-color: theme('colors.slate.50');
  }

  .suggestion-label {
    @apply text-xs font-semibold uppercase tracking-wide min-w-[4rem];
    color: theme('colors.slate.500');
  }

  .suggestion-value {
    @apply flex-1 font-medium;
  }

  .filters-grid {
    @apply grid gap-5 md:grid-cols-2;
  }

  .filters-header {
    @apply flex items-center gap-2 mb-2;
  }

  .filter-field {
    /* No specific styles needed */
  }

  .filter-label {
    @apply block text-xs uppercase tracking-widest mb-2.5 font-bold;
    color: theme('colors.slate.600');
  }

  .filter-select,
  .filter-input {
    @apply w-full rounded-lg px-4 py-3 text-sm border-0 transition-all duration-300 shadow-sm hover:shadow-md;
    background: linear-gradient(to right, theme('colors.slate.50'), theme('colors.slate.100 / 0.5'));
    color: theme('colors.slate.900');
  }

  .filter-select:focus,
  .filter-input:focus {
    @apply bg-white/95 outline-none shadow-lg;
    box-shadow: var(--input-focus-shadow);
  }

  .filter-input::placeholder {
    color: theme('colors.slate.400');
  }

  .form-actions {
    @apply flex justify-end gap-3 pt-4 border-t;
    border-color: theme('colors.slate.200 / 0.6');
  }

  .btn-reset {
    @apply rounded-lg px-6 py-2.5 text-sm font-semibold transition-all duration-300 shadow-sm hover:shadow-md border-0;
    background: linear-gradient(to right, theme('colors.slate.100 / 0.8'), theme('colors.slate.200 / 0.6'));
    color: theme('colors.slate.700');
  }

  .btn-reset:hover {
    background: linear-gradient(to right, theme('colors.slate.200 / 0.9'), theme('colors.slate.300 / 0.7'));
    transform: var(--btn-hover-lift);
  }

  .btn-search {
    @apply rounded-lg px-6 py-2.5 text-sm font-semibold text-white focus:outline-none transition-all duration-300 shadow-lg hover:shadow-xl border-0;
    background: linear-gradient(to right, theme('colors.brand.500'), theme('colors.accent.500'));
  }

  .btn-search:hover {
    background: linear-gradient(to right, theme('colors.brand.600'), theme('colors.accent.600'));
    transform: var(--btn-hover-lift);
  }

  .btn-search:focus {
    box-shadow: 0 0 0 2px theme('colors.brand.500 / 0.4'), var(--shadow-xl);
  }
</style>
