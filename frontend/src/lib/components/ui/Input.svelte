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
    value = '',
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

  // Base classes using design tokens
  const baseClasses = "w-full rounded-lg border transition-all duration-300 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed";

  // Variant classes using design tokens
  const variantClasses = {
    primary: "border-slate-300 bg-white text-slate-900 placeholder:text-slate-500 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20",
    secondary: "border-slate-200 bg-slate-50 text-slate-700 placeholder:text-slate-400 focus:border-slate-400 focus:ring-2 focus:ring-slate-400/20",
    outline: "border-brand-500 bg-transparent text-brand-600 placeholder:text-brand-400 focus:border-brand-600 focus:ring-2 focus:ring-brand-500/20",
    ghost: "border-transparent bg-transparent text-slate-700 placeholder:text-slate-400 focus:border-slate-300 focus:ring-2 focus:ring-slate-300/20",
    success: "border-success-300 bg-white text-slate-900 placeholder:text-slate-500 focus:border-success-500 focus:ring-2 focus:ring-success-500/20",
    warning: "border-warning-300 bg-white text-slate-900 placeholder:text-slate-500 focus:border-warning-500 focus:ring-2 focus:ring-warning-500/20",
    danger: "border-danger-300 bg-white text-slate-900 placeholder:text-slate-500 focus:border-danger-500 focus:ring-2 focus:ring-danger-500/20"
  };

  // Size classes using design tokens
  const sizeClasses = {
    xs: "h-(--input-height-xs) px-2 py-1 text-xs",
    sm: "h-(--input-height-sm) px-3 py-1.5 text-sm",
    md: "h-(--input-height-md) px-4 py-2 text-sm",
    lg: "h-(--input-height-lg) px-4 py-2 text-base",
    xl: "h-(--input-height-xl) px-6 py-3 text-lg"
  };

  // Computed values with $derived
  const errorClasses = $derived(error ? "border-danger-500 focus:border-danger-500 focus:ring-danger-500/20" : "");
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

<div class="input-wrapper">
  <div class="input-container {classes}" class:focused={isFocused} class:error={!!error}>
    {#if leadingIcon && leadingIconSnippet}
      <div class="input-leading-icon">
        {@render leadingIconSnippet()}
      </div>
    {/if}

    <input
      bind:this={inputElement}
      {type}
      {placeholder}
      {value}
      {required}
      {readonly}
      {disabled}
      minlength={minLength}
      maxlength={maxLength}
      {pattern}
      class="input-field"
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
  <div class="input-footer">
    {#if error}
      <p class="input-error">{error}</p>
    {:else if helperText}
      <p class="input-helper">{helperText}</p>
    {/if}

    {#if showCount && maxCount > 0}
      <p class="input-count" class:over-limit={isOverLimit}>
        {characterCount}/{maxCount}
      </p>
    {/if}
  </div>
</div>

<style>
  .input-wrapper {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .input-container {
    position: relative;
    display: flex;
    align-items: center;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .input-container.focused {
    transform: scale(1.01);
    box-shadow:
      0 0 0 3px rgba(59, 130, 246, 0.1),
      0 4px 12px rgba(59, 130, 246, 0.15);
  }

  .input-container.error.focused {
    box-shadow:
      0 0 0 3px rgba(239, 68, 68, 0.1),
      0 4px 12px rgba(239, 68, 68, 0.15);
  }

  .input-field {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    font-family: inherit;
    font-size: inherit;
    line-height: inherit;
    color: inherit;
  }

  .input-field::placeholder {
    color: inherit;
    opacity: 0.6;
  }

  .input-leading-icon,
  .input-trailing-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-slate-500);
    transition: color 0.2s ease;
  }

  .input-leading-icon {
    margin-right: var(--spacing-sm);
  }

  .input-trailing-icon {
    margin-left: var(--spacing-sm);
  }

  .input-container.focused .input-leading-icon,
  .input-container.focused .input-trailing-icon {
    color: var(--color-brand-500);
  }

  .input-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    min-height: 1rem;
  }

  .input-error {
    color: var(--color-danger-600);
    font-size: var(--text-xs);
    font-weight: 500;
    margin: 0;
  }

  .input-helper {
    color: var(--color-slate-500);
    font-size: var(--text-xs);
    margin: 0;
  }

  .input-count {
    color: var(--color-slate-400);
    font-size: var(--text-xs);
    font-weight: 500;
    margin: 0;
  }

  .input-count.over-limit {
    color: var(--color-danger-600);
  }
</style>