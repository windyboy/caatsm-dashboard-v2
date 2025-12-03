<!--
  @component Button
  Standardized button component with consistent API following design system.

  @param {ComponentVariant} variant - Button style variant
  @param {ComponentSize} size - Button size
  @param {boolean} disabled - Whether the button is disabled
  @param {boolean} loading - Whether to show loading state
  @param {boolean} ripple - Whether to enable ripple effect
  @param {string} type - Button type attribute
  @param {string} href - If provided, renders as anchor tag
  @param {boolean} fullWidth - Whether the button should take full width
  @param {string} class - Additional CSS classes
-->

<script lang="ts">
  import type { Snippet } from "svelte";
  import type { ButtonProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends ButtonProps {
    onclick?: (event: Event) => void;
    children?: Snippet;
  }

  let {
    variant = "primary",
    size = "md",
    disabled = false,
    loading = false,
    // ripple = true,
    type = "button",
    href = undefined,
    fullWidth = false,
    class: className = "",
    onclick,
    children,
  }: Props = $props();

  // Base classes using UnoCSS
  const baseClasses = "inline-flex items-center justify-center font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed";

  // Variant classes using UnoCSS
  const variantClasses = {
    primary: "bg-brand-500 text-white hover:bg-brand-600 focus:ring-brand-500",
    secondary: "bg-slate-500 text-white hover:bg-slate-600 focus:ring-slate-500",
    outline: "border-2 border-brand-500 text-brand-500 hover:bg-brand-50 focus:ring-brand-500",
    ghost: "text-brand-600 hover:bg-brand-50 focus:ring-brand-500",
    success: "bg-success-500 text-white hover:bg-success-600 focus:ring-success-500",
    warning: "bg-warning-500 text-white hover:bg-warning-600 focus:ring-warning-500",
    danger: "bg-danger-500 text-white hover:bg-danger-600 focus:ring-danger-500",
  };

  // Size classes using UnoCSS
  const sizeClasses = {
    xs: "px-2 py-1 text-xs",
    sm: "px-3 py-1.5 text-sm",
    md: "px-4 py-2 text-base",
    lg: "px-5 py-2.5 text-lg",
    xl: "px-6 py-3 text-xl",
  };

  // Computed values with $derived
  const widthClass = $derived(fullWidth ? "w-full" : "");
  const classes = $derived(
    `${baseClasses} ${variantClasses[variant!]} ${sizeClasses[size!]} ${widthClass} ${className}`.trim()
  );

  function handleClick(event: Event) {
    if (disabled || loading) {
      event.preventDefault();
      return;
    }
    onclick?.(event);
  }
</script>

{#if href}
  <a
    href={disabled || loading ? undefined : href}
    class={classes}
    class:pointer-events-none={disabled || loading}
    aria-disabled={disabled || loading}
    role="button"
    onclick={handleClick}
  >
    {#if loading}
      <span class="inline-block w-4 h-4 mr-2 border-2 border-current border-t-transparent rounded-full animate-spin"></span>
    {/if}
    {@render children?.()}
  </a>
{:else}
  <button
    {type}
    {disabled}
    aria-disabled={disabled || loading}
    class={classes}
    onclick={handleClick}
  >
    {#if loading}
      <span class="inline-block w-4 h-4 mr-2 border-2 border-current border-t-transparent rounded-full animate-spin"></span>
    {/if}
    {@render children?.()}
  </button>
{/if}
