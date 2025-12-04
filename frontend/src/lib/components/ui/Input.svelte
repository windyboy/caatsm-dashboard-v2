<script lang="ts">
  interface InputProps {
    label?: string;
    value?: string;
    placeholder?: string;
    type?: string;
    name?: string;
    disabled?: boolean;
    ariaLabel?: string;
  }

  let {
    label,
    value = $bindable(""),
    placeholder = "",
    type = "text",
    name = "",
    disabled = false,
    ariaLabel,
    ...restProps
  }: InputProps = $props();
</script>

<label class="search-field">
  {#if label}
    <span class="muted">{label}</span>
  {/if}
  <input
    {type}
    {name}
    {placeholder}
    bind:value
    {disabled}
    aria-label={ariaLabel || label || undefined}
    class="search-input"
    {...restProps}
  />
</label>
