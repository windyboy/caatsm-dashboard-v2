<script lang="ts">
  import {
    Select as SelectRoot,
    SelectContent,
    SelectItem,
    SelectTrigger,
  } from "./select/index.js";

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

  const selectedOption = $derived(options.find((opt) => opt.value === value) || null);
  let open = $state(false);

  // Watch for value changes and call onValueChange callback
  $effect(() => {
    if (onValueChange && value !== undefined && value !== null) {
      onValueChange(value);
    }
  });
</script>

<SelectRoot type="single" bind:value={value as never} {disabled} bind:open>
  <SelectTrigger {disabled} class="w-full">
    {selectedOption?.label || placeholder}
  </SelectTrigger>
  <SelectContent>
    {#each options as option}
      <SelectItem value={option.value} label={option.label} />
    {/each}
  </SelectContent>
</SelectRoot>
