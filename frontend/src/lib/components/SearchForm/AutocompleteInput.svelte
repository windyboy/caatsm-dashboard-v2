<!--
  @component AutocompleteInput
  Autocomplete input field with suggestions dropdown.
  
  @param {string} query - Current query value
  @param {(query: string) => void} oninput - Called when input value changes
  @param {Array<{value: string, type: string, label: string} | string>} suggestions - Autocomplete suggestions
  @param {boolean} showSuggestions - Whether to show suggestions dropdown
  @param {(suggestion: string | {value: string, type: string, label: string}) => void} onselect - Called when suggestion is selected
-->

<script lang="ts">
  interface Props {
    query: string;
    oninput: (query: string) => void;
    suggestions: Array<{ value: string; type: string; label: string } | string>;
    showSuggestions: boolean;
    onselect: (suggestion: string | { value: string; type: string; label: string }) => void;
  }

  let { query, oninput, suggestions, showSuggestions, onselect }: Props = $props();

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

<div class="search-field">
  <label class="search-label" for="query">Keywords</label>
  <div class="search-input-container">
    <input
      id="query"
      type="search"
      value={query}
      oninput={(e) => oninput(e.currentTarget.value)}
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
            onclick={() => onselect(suggestion)}
          >
            {#if label}
              <span class="suggestion-label">{label}:</span>
            {/if}
            <span class="suggestion-value">{value}</span>
          </button>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
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
</style>

