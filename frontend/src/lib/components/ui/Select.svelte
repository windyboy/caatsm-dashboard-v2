<!--
  @component Select
  Standardized select/dropdown component with consistent API following design system.

  @param {ComponentVariant} variant - Select variant
  @param {ComponentSize} size - Select size
  @param {SelectOption[]} options - Available options
  @param {string | number | (string | number)[]} value - Selected value(s)
  @param {string} placeholder - Placeholder text
  @param {boolean} multiple - Whether multiple selection is allowed
  @param {boolean} searchable - Whether the select is searchable
  @param {boolean} clearable - Whether the select is clearable
  @param {number} maxOptions - Maximum number of options to show
  @param {string} error - Error message
  @param {string} helperText - Helper text
  @param {string} loadingText - Loading state text
-->

<script lang="ts">
  import { tick } from "svelte";
  import type { SelectProps, SelectOption } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends SelectProps {
    onchange?: (detail: { value: string | number | (string | number)[] }) => void;
    onopen?: (event: Event) => void;
    onclose?: (event: Event) => void;
  }

  let {
    variant = 'primary',
    size = 'md',
    options = [],
    value = $bindable(),
    placeholder = 'Select an option',
    multiple = false,
    searchable = false,
    clearable = false,
    maxOptions,
    error,
    helperText,
    loadingText = 'Loading...',
    disabled = false,
    loading = false,
    class: className = '',
    onchange,
    onopen,
    onclose
  }: Props = $props();

  // Internal state
  let isOpen = $state(false);
  let searchQuery = $state('');
  let selectElement = $state<HTMLElement>();
  let dropdownElement = $state<HTMLElement>();
  
  // Generate unique ID for listbox
  const listboxId = `listbox-${Math.random().toString(36).substr(2, 9)}`;

  // Computed values with $derived
  const selectedOptions = $derived(getSelectedOptions());
  const filteredOptions = $derived(getFilteredOptions());
  const displayValue = $derived(getDisplayValue());

  function getSelectedOptions(): SelectOption[] {
    if (!value) return [];

    if (multiple && Array.isArray(value)) {
      return options.filter(option => (value as (string | number)[]).includes(option.value));
    } else if (!multiple && value !== undefined) {
      return options.filter(option => option.value === value);
    }

    return [];
  }

  function getFilteredOptions(): SelectOption[] {
    let filtered = options;

    // Filter by search query
    if (searchable && searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(option =>
        option.label.toLowerCase().includes(query) ||
        String(option.value).toLowerCase().includes(query)
      );
    }

    // Limit options if specified
    if (maxOptions && maxOptions > 0) {
      filtered = filtered.slice(0, maxOptions);
    }

    return filtered;
  }

  function getDisplayValue(): string {
    if (loading) return loadingText || 'Loading...';
    if (selectedOptions.length === 0) return placeholder || 'Select an option';

    if (multiple) {
      if (selectedOptions.length === 1) {
        return selectedOptions[0].label;
      } else {
        return `${selectedOptions.length} selected`;
      }
    } else {
      return selectedOptions[0]?.label || (placeholder || 'Select an option');
    }
  }

  function toggleDropdown() {
    if (disabled || loading) return;
    isOpen = !isOpen;

    if (isOpen) {
      onopen?.(new Event('open'));
      // Focus search input if searchable
      if (searchable) {
        tick().then(() => {
          const searchInput = dropdownElement?.querySelector('input');
          searchInput?.focus();
        });
      }
    } else {
      onclose?.(new Event('close'));
      searchQuery = '';
    }
  }

  function selectOption(option: SelectOption) {
    if (option.disabled) return;

    let newValue: string | number | (string | number)[];

    if (multiple) {
      const currentValues = Array.isArray(value) ? value : [];
      const isSelected = currentValues.includes(option.value);

      if (isSelected) {
        newValue = currentValues.filter(v => v !== option.value);
      } else {
        newValue = [...currentValues, option.value];
      }
    } else {
      newValue = option.value;
      isOpen = false;
    }

    value = newValue;
    onchange?.({ value: newValue });
  }

  function clearSelection() {
    if (!clearable) return;
    const newValue = multiple ? [] : ('' as any);
    value = newValue;
    onchange?.({ value: newValue });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (disabled || loading) return;

    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        if (!isOpen) {
          toggleDropdown();
        }
        break;
      case 'Escape':
        if (isOpen) {
          isOpen = false;
          selectElement?.focus();
        }
        break;
      case 'ArrowDown':
        if (!isOpen) {
          toggleDropdown();
        }
        break;
    }
  }

  // Close dropdown when clicking outside
  function handleClickOutside(event: MouseEvent) {
    if (isOpen && !selectElement?.contains(event.target as Node)) {
      isOpen = false;
      searchQuery = '';
    }
  }

  // Base classes using design tokens
  const baseClasses = "relative w-full rounded-lg border transition-all duration-300 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed";

  // Variant classes using design tokens
  const variantClasses = {
    primary: "border-slate-300 bg-white text-slate-900 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20",
    secondary: "border-slate-200 bg-slate-50 text-slate-700 focus:border-slate-400 focus:ring-2 focus:ring-slate-400/20",
    outline: "border-brand-500 bg-transparent text-brand-600 focus:border-brand-600 focus:ring-2 focus:ring-brand-500/20",
    ghost: "border-transparent bg-transparent text-slate-700 focus:border-slate-300 focus:ring-2 focus:ring-slate-300/20",
    success: "border-success-300 bg-white text-slate-900 focus:border-success-500 focus:ring-2 focus:ring-success-500/20",
    warning: "border-warning-300 bg-white text-slate-900 focus:border-warning-500 focus:ring-2 focus:ring-warning-500/20",
    danger: "border-danger-300 bg-white text-slate-900 focus:border-danger-500 focus:ring-2 focus:ring-danger-500/20"
  };

  // Size classes using design tokens
  const sizeClasses = {
    xs: "h-[var(--input-height-xs)] px-2 py-1 text-xs",
    sm: "h-[var(--input-height-sm)] px-3 py-1.5 text-sm",
    md: "h-[var(--input-height-md)] px-4 py-2 text-sm",
    lg: "h-[var(--input-height-lg)] px-4 py-2 text-base",
    xl: "h-[var(--input-height-xl)] px-6 py-3 text-lg"
  };

  // Computed classes with $derived
  const errorClasses = $derived(error ? "border-danger-500 focus:border-danger-500 focus:ring-danger-500/20" : "");
  const selectClasses = $derived(`${baseClasses} ${variantClasses[variant!]} ${sizeClasses[size!]} ${errorClasses} ${className}`.trim());
</script>

<svelte:window onclick={handleClickOutside} />

<div class="select-wrapper">
  <div
    bind:this={selectElement}
    class={selectClasses}
    class:open={isOpen}
    class:error={!!error}
    onclick={toggleDropdown}
    onkeydown={handleKeydown}
    role="combobox"
    aria-expanded={isOpen}
    aria-controls={listboxId}
    aria-haspopup="listbox"
    tabindex={disabled ? -1 : 0}
  >
    <div class="select-content">
      <span class="select-value" class:placeholder={!selectedOptions.length}>
        {displayValue}
      </span>

      <div class="select-actions">
        {#if clearable && selectedOptions.length > 0 && !loading}
          <button
            type="button"
            class="select-clear"
            onclick={(e: Event) => { e.stopPropagation(); clearSelection(); }}
            aria-label="Clear selection"
          >
            ×
          </button>
        {/if}

        <div class="select-arrow" class:rotated={isOpen}>
          ▼
        </div>
      </div>
    </div>
  </div>

  {#if isOpen}
    <div
      bind:this={dropdownElement}
      id={listboxId}
      class="select-dropdown"
      role="listbox"
      aria-multiselectable={multiple}
    >
      {#if searchable}
        <div class="select-search">
          <input
            type="text"
            placeholder="Search..."
            bind:value={searchQuery}
            class="select-search-input"
          />
        </div>
      {/if}

      <div class="select-options">
        {#each filteredOptions as option (option.value)}
          {@const isSelected = selectedOptions.some(selected => selected.value === option.value)}
          <div
            class="select-option"
            class:selected={isSelected}
            class:disabled={option.disabled}
            onclick={() => selectOption(option)}
            onkeydown={(e: KeyboardEvent) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                selectOption(option);
              }
            }}
            role="option"
            aria-selected={isSelected}
            tabindex={option.disabled ? -1 : 0}
          >
            {#if multiple}
              <div class="select-checkbox">
                {#if isSelected}
                  ✓
                {/if}
              </div>
            {/if}
            <span class="select-option-label">{option.label}</span>
          </div>
        {:else}
          <div class="select-empty">
            No options available
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Helper text and error messages -->
  <div class="select-footer">
    {#if error}
      <p class="select-error">{error}</p>
    {:else if helperText}
      <p class="select-helper">{helperText}</p>
    {/if}
  </div>
</div>

<style>
  .select-wrapper {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .select-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
  }

  .select-value {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .select-value.placeholder {
    color: var(--color-slate-500);
    opacity: 0.6;
  }

  .select-actions {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    margin-left: var(--spacing-sm);
  }

  .select-clear {
    background: none;
    border: none;
    color: var(--color-slate-400);
    cursor: pointer;
    font-size: 1.2em;
    line-height: 1;
    padding: 0;
    transition: color 0.2s ease;
  }

  .select-clear:hover {
    color: var(--color-danger-500);
  }

  .select-arrow {
    color: var(--color-slate-400);
    font-size: 0.8em;
    transition: transform 0.2s ease;
    user-select: none;
  }

  .select-arrow.rotated {
    transform: rotate(180deg);
  }

  .select-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    z-index: 1000;
    background: white;
    border: 1px solid var(--color-slate-200);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    max-height: 200px;
    overflow: hidden;
    margin-top: var(--spacing-xs);
  }

  .select-search {
    padding: var(--spacing-sm);
    border-bottom: 1px solid var(--color-slate-100);
  }

  .select-search-input {
    width: 100%;
    border: 1px solid var(--color-slate-200);
    border-radius: var(--radius-md);
    padding: var(--spacing-sm);
    font-size: var(--text-sm);
  }

  .select-options {
    max-height: 160px;
    overflow-y: auto;
  }

  .select-option {
    padding: var(--spacing-sm) var(--spacing-md);
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    transition: background-color 0.2s ease;
  }

  .select-option:hover:not(.disabled) {
    background-color: var(--color-slate-50);
  }

  .select-option.selected {
    background-color: var(--color-brand-50);
    color: var(--color-brand-600);
  }

  .select-option.disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  .select-checkbox {
    width: 1rem;
    height: 1rem;
    border: 1px solid var(--color-slate-300);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8em;
    color: var(--color-brand-600);
  }

  .select-empty {
    padding: var(--spacing-md);
    color: var(--color-slate-500);
    text-align: center;
    font-style: italic;
  }

  .select-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    min-height: 1rem;
  }

  .select-error {
    color: var(--color-danger-600);
    font-size: var(--text-xs);
    font-weight: 500;
    margin: 0;
  }

  .select-helper {
    color: var(--color-slate-500);
    font-size: var(--text-xs);
    margin: 0;
  }
</style>