<!--
  @component SearchForm
  Search form for filtering telegram messages with autocomplete and filters.

  Features:
  - Keyword search with autocomplete
  - Type, priority, and time range filters
  - Debounced autocomplete (500ms)
  - Form validation and reset functionality

  @prop {(detail: SearchParams) => void} [onsearch] - Called when search is performed
  @prop {() => void} [onreset] - Called when form is reset
-->

<script lang="ts">
  import { onDestroy } from "svelte";
  import { autocomplete } from "../services/api";
  import { createLogger } from "../utils/logger";
  import type { SearchParams } from "../services/api";
  import { SEARCH_CONFIG } from "../constants";

  const logger = createLogger("SearchForm");

  interface Props {
    onsearch?: (detail: SearchParams) => void;
    onreset?: () => void;
    initialParams?: {
      query?: string;
      type?: string;
      priority?: string;
      start_time?: string;
      end_time?: string;
    };
  }

  let { onsearch, onreset, initialParams }: Props = $props();

  let query = $state("");
  let type = $state("");
  let priority = $state("");
  let start_time = $state("");
  let end_time = $state("");

  // Sync state when initialParams changes (including initial mount)
  $effect(() => {
    if (initialParams) {
      query = initialParams.query || "";
      type = initialParams.type || "";
      priority = initialParams.priority || "";
      start_time = initialParams.start_time || "";
      end_time = initialParams.end_time || "";
    }
  });
  let suggestions = $state<Array<{ value: string; type: string; label: string } | string>>([]);
  let showSuggestions = $state(false);
  let autocompleteTimeout = $state<ReturnType<typeof setTimeout> | null>(null);
  let error = $state<string | null>(null);
  let abortController = $state<AbortController | null>(null);

  onDestroy(() => {
    if (autocompleteTimeout !== null) {
      clearTimeout(autocompleteTimeout);
      autocompleteTimeout = null;
    }
    if (abortController) {
      abortController.abort();
      abortController = null;
    }
  });

  // Convert datetime-local (local time) to ISO string (UTC) - API expects UTC timestamps
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
    // Cancel previous request
    if (abortController) {
      abortController.abort();
    }

    if (autocompleteTimeout) {
      clearTimeout(autocompleteTimeout);
    }

    if (query.length < SEARCH_CONFIG.AUTOCOMPLETE_MIN_LENGTH) {
      suggestions = [];
      showSuggestions = false;
      return;
    }

    autocompleteTimeout = setTimeout(async () => {
      abortController = new AbortController();
      const currentQuery = query; // Capture current value

      try {
        const result = await autocomplete(query, 5, abortController.signal);

        // Only update if query hasn't changed and request wasn't aborted
        if (currentQuery === query && !abortController.signal.aborted) {
          // Handle both old format (string[]) and new format (AutocompleteSuggestion[])
          if (result.suggestions.length > 0 && typeof result.suggestions[0] === "string") {
            // Old format: convert to new format
            suggestions = (result.suggestions as string[]).map((s) => ({
              value: s,
              type: "text",
              label: "",
            }));
          } else {
            suggestions = result.suggestions as Array<{
              value: string;
              type: string;
              label: string;
            }>;
          }
          showSuggestions = suggestions.length > 0;
        }
      } catch (error) {
        // Ignore AbortError - it's expected when request is cancelled
        if (error instanceof Error && error.name !== "AbortError") {
          logger.error("Autocomplete failed", error, {
            query: currentQuery,
            queryLength: currentQuery.length,
          });
        }
        // Only clear suggestions if request wasn't aborted
        if (!abortController.signal.aborted) {
          suggestions = [];
          showSuggestions = false;
        }
      }
    }, SEARCH_CONFIG.AUTOCOMPLETE_DEBOUNCE_MS);
  }

  function handleSearch() {
    error = null;

    const startISO = toIsoString(start_time);
    const endISO = toIsoString(end_time);

    // Validate time range if both dates are provided
    if (startISO && endISO) {
      const start = new Date(startISO);
      const end = new Date(endISO);

      // Check if start time is before end time
      if (start >= end) {
        error = "Start time must be before end time";
        return;
      }

      // Check if time range exceeds maximum (90 days)
      const diffMs = end.getTime() - start.getTime();
      const diffDays = diffMs / (1000 * 60 * 60 * 24);

      if (diffDays > SEARCH_CONFIG.MAX_TIME_RANGE_DAYS) {
        error = `Time range cannot exceed ${SEARCH_CONFIG.MAX_TIME_RANGE_DAYS} days`;
        return;
      }
    }

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
    error = null;
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
      <h2 class="search-title">Search Telegrams</h2>
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
    {#if error}
      <div class="error-message">
        <div class="error-icon">⚠️</div>
        <span class="error-text">{error}</span>
      </div>
    {/if}
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
      <button type="submit" class="btn-search"> Search </button>
    </div>
  </form>
</div>

<style>
  .search-form {
    border-radius: 0.5rem;
    border: 0;
    padding: 2rem;
    background: var(--card-glow-bg);
    backdrop-filter: var(--card-glow-backdrop);
    box-shadow: var(--card-glow-shadow);
    border: var(--card-glow-border);
    transition: var(--card-glow-transition);
  }

  .search-header {
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid rgba(191, 219, 254, 0.4);
  }

  .search-header-content {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .search-icon {
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(191, 219, 254, 0.5);
    background: linear-gradient(to right, #dbeafe, #f3e8ff);
  }

  .search-title {
    font-size: 1.25rem;
    line-height: 1.75rem;
    font-weight: 700;
    letter-spacing: -0.015em;
    color: #1e293b;
    background: linear-gradient(to right, #2563eb, #9333ea);
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .search-form-fields {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .search-label {
    display: block;
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 600;
    margin-bottom: 0.625rem;
    color: #334155;
  }

  .search-input-container {
    position: relative;
  }

  .search-input {
    width: 100%;
    border-radius: 0.5rem;
    padding: 0.75rem 1.25rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    color: #0f172a;
    border: 0;
    background: linear-gradient(to right, #f8fafc, rgba(241, 245, 249, 0.5));
    transition: all 0.3s ease;
    box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.08);
  }

  .search-input:hover {
    box-shadow:
      0 4px 6px -1px rgba(15, 23, 42, 0.1),
      0 2px 4px -2px rgba(15, 23, 42, 0.1);
  }

  .search-input:focus {
    background: rgba(255, 255, 255, 0.95);
    outline: none;
    box-shadow: var(--input-focus-shadow);
  }

  .search-input::placeholder {
    color: #94a3b8;
  }

  .suggestions-dropdown {
    position: absolute;
    z-index: 10;
    width: 100%;
    margin-top: 0.25rem;
    background: #ffffff;
    border-radius: 0.5rem;
    box-shadow:
      0 10px 15px -3px rgba(15, 23, 42, 0.1),
      0 4px 6px -4px rgba(15, 23, 42, 0.1);
    border: 1px solid #e2e8f0;
    max-height: 12rem;
    overflow-y: auto;
  }

  .suggestion-item {
    width: 100%;
    text-align: left;
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: #334155;
    background: transparent;
    border: 0;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .suggestion-item:hover {
    background-color: #f8fafc;
  }

  .suggestion-label {
    font-size: 0.75rem;
    line-height: 1rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    min-width: 4rem;
    color: #64748b;
  }

  .suggestion-value {
    flex: 1;
    font-weight: 500;
  }

  .filters-grid {
    display: grid;
    gap: 1.25rem;
  }

  @media (min-width: 768px) {
    .filters-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  .filters-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .filter-label {
    display: block;
    font-size: 0.75rem;
    line-height: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    margin-bottom: 0.625rem;
    font-weight: 700;
    color: #475569;
  }

  .filter-select,
  .filter-input {
    width: 100%;
    border-radius: 0.5rem;
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    border: 0;
    background: linear-gradient(to right, #f8fafc, rgba(241, 245, 249, 0.5));
    color: #0f172a;
    transition: all 0.3s ease;
    box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.08);
  }

  .filter-select:hover,
  .filter-input:hover {
    box-shadow:
      0 4px 6px -1px rgba(15, 23, 42, 0.1),
      0 2px 4px -2px rgba(15, 23, 42, 0.1);
  }

  .filter-select:focus,
  .filter-input:focus {
    background: rgba(255, 255, 255, 0.95);
    outline: none;
    box-shadow: var(--input-focus-shadow);
  }

  .filter-input::placeholder {
    color: #94a3b8;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(254, 202, 202, 0.6);
    background: linear-gradient(to right, #fef2f2, #fff7ed);
    color: #b91c1c;
  }

  .error-icon {
    flex-shrink: 0;
    font-size: 1.125rem;
  }

  .error-text {
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 600;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    padding-top: 1rem;
    border-top: 1px solid rgba(226, 232, 240, 0.6);
  }

  .btn-reset {
    border-radius: 0.5rem;
    padding: 0.625rem 1.5rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 600;
    border: 0;
    color: #334155;
    background: linear-gradient(to right, rgba(241, 245, 249, 0.8), rgba(226, 232, 240, 0.6));
    box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.06);
    transition: all 0.3s ease;
    cursor: pointer;
  }

  .btn-reset:hover {
    background: linear-gradient(to right, rgba(226, 232, 240, 0.9), rgba(203, 213, 225, 0.7));
    box-shadow:
      0 4px 6px -1px rgba(15, 23, 42, 0.1),
      0 2px 4px -2px rgba(15, 23, 42, 0.1);
    transform: var(--btn-hover-lift);
  }

  .btn-search {
    border-radius: 0.5rem;
    padding: 0.625rem 1.5rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    font-weight: 600;
    color: #ffffff;
    border: 0;
    background: linear-gradient(to right, #3b82f6, #a855f7);
    box-shadow:
      0 10px 15px -3px rgba(15, 23, 42, 0.2),
      0 4px 6px -4px rgba(15, 23, 42, 0.1);
    transition: all 0.3s ease;
    cursor: pointer;
    outline: none;
  }

  .btn-search:hover {
    background: linear-gradient(to right, #2563eb, #9333ea);
    box-shadow:
      0 12px 18px -3px rgba(15, 23, 42, 0.25),
      0 4px 6px -4px rgba(15, 23, 42, 0.15);
    transform: var(--btn-hover-lift);
  }

  .btn-search:focus {
    outline: none;
    box-shadow:
      0 0 0 2px rgba(59, 130, 246, 0.4),
      var(--shadow-xl);
  }
</style>
