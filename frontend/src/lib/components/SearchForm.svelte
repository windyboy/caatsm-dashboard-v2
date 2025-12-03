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
  import type { SearchParams } from "../services/api";
  import { useSearchForm } from "../composables/useSearchForm.svelte.ts";
  import { useAutocomplete } from "../composables/useAutocomplete.svelte.ts";
  import AutocompleteInput from "./SearchForm/AutocompleteInput.svelte";
  import FilterFields from "./SearchForm/FilterFields.svelte";
  import FormActions from "./SearchForm/FormActions.svelte";

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

  const form = useSearchForm(() => initialParams);
  const autocomplete = useAutocomplete();

  function handleQueryInput(query: string) {
    form.query = query;
    autocomplete.handleAutocomplete(query);
  }

  function handleFilterChange(field: string, value: string) {
    if (field === "type") {
      form.type = value;
    } else if (field === "priority") {
      form.priority = value;
    } else if (field === "start_time") {
      form.start_time = value;
    } else if (field === "end_time") {
      form.end_time = value;
    }
  }

  function handleSuggestionSelect(suggestion: string | { value: string; type: string; label: string }) {
    const value = typeof suggestion === "string" ? suggestion : suggestion.value;
    form.query = value;
    autocomplete.clearSuggestions();
    handleSearch();
  }

  function handleSearch() {
    if (!form.validate()) {
      return;
    }

    onsearch?.(form.getSearchParams());
  }

  function handleReset() {
    form.reset();
    autocomplete.clearSuggestions();
    onreset?.();
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
    <AutocompleteInput
      query={form.query}
      oninput={handleQueryInput}
      suggestions={autocomplete.suggestions}
      showSuggestions={autocomplete.showSuggestions}
      onselect={handleSuggestionSelect}
    />
    <FilterFields
      type={form.type}
      priority={form.priority}
      start_time={form.start_time}
      end_time={form.end_time}
      onchange={handleFilterChange}
    />
    <FormActions error={form.error} onreset={handleReset} />
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
</style>
