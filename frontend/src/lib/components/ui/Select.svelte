<script lang="ts">
  import {
    Select as SelectRoot,
    SelectContent,
    SelectItem,
    SelectTrigger,
  } from "./select/index.js";

  /**
   * Select option type
   */
  export type SelectOption = {
    value: string;
    label: string;
  };

  /**
   * Select component wrapper that simplifies shadcn Select usage with an options-based API.
   * Automatically handles option rendering and value change callbacks.
   * Use this component for simple select dropdowns. For complex selects with groups,
   * use the base Select components directly from "./select/index.js".
   *
   * @example
   * ```svelte
   * <Select
   *   options={[
   *     { value: "option1", label: "Option 1" },
   *     { value: "option2", label: "Option 2" }
   *   ]}
   *   bind:value={selectedValue}
   *   onValueChange={(value) => console.log(value)}
   * />
   * ```
   */
  interface SelectProps {
    /** Array of select options */
    options: SelectOption[];
    /** Selected value (bindable) */
    value?: string;
    /** Placeholder text when no option is selected */
    placeholder?: string;
    /** Whether the select is disabled */
    disabled?: boolean;
    /** Callback fired when value changes (only fires on user interaction, not initial mount) */
    onValueChange?: (value: string) => void;
  }

  let {
    options = [],
    value = $bindable(""),
    placeholder = "Select...",
    disabled = false,
    onValueChange,
  }: SelectProps = $props();

  const selectedOption = $derived(options.find((opt) => opt.value === value) || null);
  let open = $state(false);
  let previousValue = $state<string | undefined>(undefined);
  let isInitialMount = $state(true);

  // Watch for value changes and call onValueChange callback only when value actually changes
  $effect(() => {
    if (isInitialMount) {
      isInitialMount = false;
      previousValue = value;
      return;
    }

    if (onValueChange && value !== undefined && value !== null && value !== previousValue) {
      onValueChange(value);
      previousValue = value;
    }
  });
</script>

<SelectRoot type="single" bind:value={value as never} {disabled} bind:open>
  <SelectTrigger {disabled} class="w-full">
    {selectedOption?.label || placeholder}
  </SelectTrigger>
  <SelectContent>
    {#each options as option (option.value)}
      <SelectItem value={option.value} label={option.label} />
    {/each}
  </SelectContent>
</SelectRoot>
