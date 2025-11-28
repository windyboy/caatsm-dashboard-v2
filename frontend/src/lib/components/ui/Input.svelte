<!--
  @component Input
  Standardized input component with consistent API following design system.

  @param {InputProps['type']} type - Input type
  @param {ComponentVariant} variant - Input variant
  @param {ComponentSize} size - Input size
  @param {string} placeholder - Placeholder text
  @param {string | number} value - Input value
  @param {boolean} required - Whether input is required
  @param {boolean} readonly - Whether input is readonly
  @param {number} minLength - Minimum length
  @param {number} maxLength - Maximum length
  @param {string} pattern - Validation pattern
  @param {string} error - Error message
  @param {string} helperText - Helper text
  @param {boolean} leadingIcon - Whether to show leading icon slot
  @param {boolean} trailingIcon - Whether to show trailing icon slot
  @param {boolean} showCount - Whether to show character count
-->

<script lang="ts">
  import type { Snippet } from "svelte";
  import type { InputProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends InputProps {
    oninput?: (event: Event) => void;
    onchange?: (event: Event) => void;
    onfocus?: (event: Event) => void;
    onblur?: (event: Event) => void;
    leadingIconSnippet?: Snippet;
    trailingIconSnippet?: Snippet;
  }

  let {
    type = 'text',
    variant = 'primary',
    size = 'md',
    placeholder = '',
    value = $bindable(''),
    required = false,
    readonly = false,
    disabled = false,
    minLength,
    maxLength,
    pattern,
    error,
    helperText,
    leadingIcon = false,
    trailingIcon = false,
    showCount = false,
    class: className = '',
    oninput,
    onchange,
    onfocus,
    onblur,
    leadingIconSnippet,
    trailingIconSnippet
  }: Props = $props();

  // Internal state
  let inputElement = $state<HTMLInputElement>();
  let isFocused = $state(false);

  // Base classes using Daisy UI
  const baseClasses = "input";

  // Variant classes using Daisy UI
  const variantClasses = {
    primary: "input-bordered",
    secondary: "input-bordered input-ghost",
    outline: "input-bordered input-primary",
    ghost: "input-ghost",
    success: "input-bordered input-success",
    warning: "input-bordered input-warning",
    danger: "input-bordered input-error"
  };

  // Size classes using Daisy UI
  const sizeClasses = {
    xs: "input-xs",
    sm: "input-sm",
    md: "input-md",
    lg: "input-lg",
    xl: "input-xl"
  };

  // Computed values with $derived
  const errorClasses = $derived(error ? "input-error" : "");
  const classes = $derived(`${baseClasses} ${variantClasses[variant!]} ${sizeClasses[size!]} ${errorClasses} ${className}`.trim());
  const characterCount = $derived(showCount && typeof value === 'string' ? value.length : 0);
  const maxCount = $derived(maxLength || 0);
  const isOverLimit = $derived(showCount && maxCount > 0 && characterCount > maxCount);

  function handleInput(event: Event) {
    oninput?.(event);
  }

  function handleChange(event: Event) {
    onchange?.(event);
  }

  function handleFocus(event: Event) {
    isFocused = true;
    onfocus?.(event);
  }

  function handleBlur(event: Event) {
    isFocused = false;
    onblur?.(event);
  }

  // Expose focus method
  export function focus() {
    inputElement?.focus();
  }

  // Expose blur method
  export function blur() {
    inputElement?.blur();
  }
</script>

<div class="form-control">
  {#if leadingIcon && leadingIconSnippet}
    <div class="input-leading-icon">
      {@render leadingIconSnippet()}
    </div>
  {/if}
  <input
    bind:this={inputElement}
    {type}
    {placeholder}
    bind:value={value}
    {required}
    {readonly}
    {disabled}
    minlength={minLength}
    maxlength={maxLength}
    {pattern}
    class={classes}
    oninput={handleInput}
    onchange={handleChange}
    onfocus={handleFocus}
    onblur={handleBlur}
  />
  {#if trailingIcon && trailingIconSnippet}
    <div class="input-trailing-icon">
      {@render trailingIconSnippet()}
    </div>
  {/if}
</div>

  <!-- Helper text and error messages -->
  {#if error}
    <div class="label">
      <span class="label-text-alt text-error">{error}</span>
    </div>
  {:else if helperText}
    <div class="label">
      <span class="label-text-alt">{helperText}</span>
    </div>
  {/if}

  {#if showCount && maxCount > 0}
    <div class="label">
      <span class="label-text-alt {isOverLimit ? 'text-error' : ''}">
        {characterCount}/{maxCount}
      </span>
    </div>
  {/if}
</div>
