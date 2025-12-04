<script lang="ts">
  // @ts-nocheck - Disable TypeScript checking due to Melt UI type issues
  import { createSelect, melt } from "@melt-ui/svelte";
  import { browser } from "$app/environment";
  import { onMount } from "svelte";
  import { cn } from "./utils";
  import type { Snippet } from "svelte";

  export type SelectOption = {
    value: string;
    label: string;
  };

  interface SelectProps {
    options: SelectOption[];
    value?: string;
    placeholder?: string;
    disabled?: boolean;
    onValueChange?: (value: string) => void;
  }

  let {
    options = [],
    value = $bindable(""),
    placeholder = "Select...",
    disabled = false,
    onValueChange,
  }: SelectProps = $props();

  const isDisabled = $derived(disabled);

  let mounted = $state(false);
  // Store the entire select instance
  let selectInstance = $state<ReturnType<typeof createSelect> | null>(null);

  // Extract stores to top-level variables for template access - must use $state for reactivity
  let selectedLabelStore = $state<
    ReturnType<typeof createSelect>["states"]["selectedLabel"] | null
  >(null);
  let openStore = $state<ReturnType<typeof createSelect>["states"]["open"] | null>(null);
  let selectedStore = $state<ReturnType<typeof createSelect>["states"]["selected"] | null>(null);

  const selectedOption = $derived(options.find((opt) => opt.value === value) || null);

  // Safely derive the selected label from the store
  const selectedLabel = $derived.by(() => {
    if (selectedLabelStore) {
      return $selectedLabelStore;
    }
    return null;
  });

  // Safely derive the open state from the store
  const isOpen = $derived.by(() => {
    if (openStore) {
      return $openStore;
    }
    return false;
  });

  onMount(() => {
    if (browser) {
      // @ts-ignore - Melt UI type issues
      const instance = createSelect({
        disabled: isDisabled,
        // @ts-ignore - Melt UI type issues with onSelectedChange
        onSelectedChange: (change) => {
          const next = change.next;
          if (next) {
            // next is ListboxOption<unknown>, we need to extract the value
            const optionValue =
              typeof next === "object" && "value" in next ? String(next.value) : String(next);
            value = optionValue;
            onValueChange?.(optionValue);
          } else {
            value = "";
            onValueChange?.("");
          }
        },
      });

      // Set initial value if provided
      if (selectedOption) {
        instance.states.selected.set({ value: selectedOption.value, label: selectedOption.label });
      }

      // @ts-ignore - Melt UI type issues
      selectInstance = instance;
      // @ts-ignore - Melt UI type issues
      selectedLabelStore = instance.states.selectedLabel;
      // @ts-ignore - Melt UI type issues
      openStore = instance.states.open;
      // @ts-ignore - Melt UI type issues
      selectedStore = instance.states.selected;
      mounted = true;
    }
  });
</script>

<div class="select-wrapper">
  {#if mounted && selectInstance && selectedLabelStore && openStore}
    <!-- @ts-ignore - Melt UI type issues -->
    <button
      use:melt={selectInstance.elements.trigger}
      class={cn("select-trigger", disabled && "select-trigger--disabled")}
      type="button"
    >
      <span class="select-value">
        {selectedLabel ?? selectedOption?.label ?? placeholder}
      </span>
      <span class="select-icon" aria-hidden="true">
        <svg
          width="15"
          height="15"
          viewBox="0 0 15 15"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M4.93179 5.43179C4.75605 5.60753 4.75605 5.89245 4.93179 6.06819C5.10753 6.24392 5.39245 6.24392 5.56819 6.06819L7.49999 4.13638L9.43179 6.06819C9.60753 6.24392 9.89245 6.24392 10.0682 6.06819C10.2439 5.89245 10.2439 5.60753 10.0682 5.43179L7.81819 3.18179C7.73379 3.09738 7.61933 3.04999 7.49999 3.04999C7.38064 3.04999 7.26618 3.09738 7.18179 3.18179L4.93179 5.43179ZM10.0682 9.56819C10.2439 9.39245 10.2439 9.10753 10.0682 8.93179C9.89245 8.75606 9.60753 8.75606 9.43179 8.93179L7.49999 10.8636L5.56819 8.93179C5.39245 8.75606 5.10753 8.75606 4.93179 8.93179C4.75605 9.10753 4.75605 9.39245 4.93179 9.56819L7.18179 11.8182C7.35753 11.9939 7.64245 11.9939 7.81819 11.8182L10.0682 9.56819Z"
            fill="currentColor"
            fill-rule="evenodd"
            clip-rule="evenodd"
          ></path>
        </svg>
      </span>
    </button>

    {#if isOpen}
      <div {...melt(selectInstance.elements.menu)} class="select-menu">
        {#each options as opt}
          <div
            {...melt(selectInstance.elements.option({ value: opt.value, label: opt.label }))}
            class={cn("select-option", value === opt.value && "select-option--selected")}
          >
            {opt.label}
          </div>
        {/each}
      </div>
    {/if}
  {:else}
    <!-- SSR fallback -->
    <button
      class={cn("select-trigger", disabled && "select-trigger--disabled")}
      type="button"
      {disabled}
    >
      <span class="select-value">
        {selectedOption?.label || placeholder}
      </span>
      <span class="select-icon" aria-hidden="true">
        <svg
          width="15"
          height="15"
          viewBox="0 0 15 15"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M4.93179 5.43179C4.75605 5.60753 4.75605 5.89245 4.93179 6.06819C5.10753 6.24392 5.39245 6.24392 5.56819 6.06819L7.49999 4.13638L9.43179 6.06819C9.60753 6.24392 9.89245 6.24392 10.0682 6.06819C10.2439 5.89245 10.2439 5.60753 10.0682 5.43179L7.81819 3.18179C7.73379 3.09738 7.61933 3.04999 7.49999 3.04999C7.38064 3.04999 7.26618 3.09738 7.18179 3.18179L4.93179 5.43179ZM10.0682 9.56819C10.2439 9.39245 10.2439 9.10753 10.0682 8.93179C9.89245 8.75606 9.60753 8.75606 9.43179 8.93179L7.49999 10.8636L5.56819 8.93179C5.39245 8.75606 5.10753 8.75606 4.93179 8.93179C4.75605 9.10753 4.75605 9.39245 4.93179 9.56819L7.18179 11.8182C7.35753 11.9939 7.64245 11.9939 7.81819 11.8182L10.0682 9.56819Z"
            fill="currentColor"
            fill-rule="evenodd"
            clip-rule="evenodd"
          ></path>
        </svg>
      </span>
    </button>
  {/if}
</div>

<style>
  .select-wrapper {
    position: relative;
    width: 100%;
  }

  .select-trigger {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: 1px solid #e4e4e7;
    border-radius: 6px;
    background: #fff;
    color: #18181b;
    font-size: 0.875rem;
    cursor: pointer;
    transition:
      border-color 0.2s ease,
      background-color 0.2s ease;
  }

  .select-trigger:hover:not(.select-trigger--disabled) {
    border-color: #a1a1aa;
    background: #fafafa;
  }

  .select-trigger--disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .select-value {
    flex: 1;
    text-align: left;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .select-icon {
    display: flex;
    align-items: center;
    margin-left: 0.5rem;
    color: #71717a;
    transition: transform 0.2s ease;
  }

  /* Used by Melt UI - data-state attribute is added dynamically */
  .select-trigger[data-state="open"] .select-icon {
    transform: rotate(180deg);
  }

  .select-menu {
    position: absolute;
    z-index: 50;
    width: 100%;
    margin-top: 0.25rem;
    border: 1px solid #e4e4e7;
    border-radius: 6px;
    background: #fff;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
    max-height: 300px;
    overflow-y: auto;
  }

  .select-option {
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    color: #18181b;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .select-option:hover {
    background: #f4f4f5;
  }

  .select-option--selected {
    background: #f4f4f5;
    color: #09090b;
    font-weight: 500;
  }

  /* Used by Melt UI - data-highlighted attribute is added dynamically */
  .select-option[data-highlighted] {
    background: #f4f4f5;
  }
</style>
