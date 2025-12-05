<script lang="ts">
  import { Input as ShadcnInput } from "./input/index.js";
  import { Label } from "./label/index.js";

  interface InputProps {
    label?: string;
    value?: string;
    placeholder?: string;
    type?: string;
    name?: string;
    disabled?: boolean;
    ariaLabel?: string;
    class?: string;
  }

  let {
    label,
    value = $bindable(""),
    placeholder = "",
    type = "text",
    name = "",
    disabled = false,
    ariaLabel,
    class: className,
    ...restProps
  }: InputProps = $props();
</script>

<div class="flex flex-col gap-1.5">
  {#if label}
    <Label for={name || undefined} class="text-muted-foreground text-sm">
      {label}
    </Label>
  {/if}
  <ShadcnInput
    {type}
    {name}
    {placeholder}
    bind:value
    {disabled}
    aria-label={ariaLabel || label || undefined}
    class={className}
    {...restProps}
  />
</div>
