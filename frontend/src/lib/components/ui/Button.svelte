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
  import LoadingSpinner from "./LoadingSpinner.svelte";
  import type { ButtonProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends ButtonProps {
    onclick?: (event: Event) => void;
    children?: Snippet;
  }

  let {
    variant = 'primary',
    size = 'md',
    disabled = false,
    loading = false,
    ripple = true,
    type = 'button',
    href = undefined,
    fullWidth = false,
    class: className = '',
    onclick,
    children
  }: Props = $props();

  // Base classes using design tokens
  const baseClasses = "inline-flex items-center justify-center font-semibold rounded-lg transition-all duration-300 focus:outline-none focus:ring-2 disabled:opacity-50 disabled:cursor-not-allowed";

  // Variant classes using design tokens
  const variantClasses = {
    primary: "bg-gradient-to-r from-brand-500 to-accent-500 text-white hover:from-brand-600 hover:to-accent-600 focus:ring-brand-500/40 shadow-lg hover:shadow-xl",
    secondary: "bg-gradient-to-r from-slate-100 to-slate-200 text-slate-700 hover:from-slate-200 hover:to-slate-300 focus:ring-slate-500/40 shadow-sm hover:shadow-md",
    outline: "border-2 border-brand-500 text-brand-600 bg-transparent hover:bg-brand-50 focus:ring-brand-500/40",
    ghost: "text-slate-700 hover:text-brand-600 hover:bg-brand-50/80 focus:ring-brand-500/40",
    success: "bg-gradient-to-r from-success-500 to-success-600 text-white hover:from-success-600 hover:to-success-700 focus:ring-success-500/40 shadow-lg hover:shadow-xl",
    warning: "bg-gradient-to-r from-warning-500 to-warning-600 text-white hover:from-warning-600 hover:to-warning-700 focus:ring-warning-500/40 shadow-lg hover:shadow-xl",
    danger: "bg-gradient-to-r from-danger-500 to-danger-600 text-white hover:from-danger-600 hover:to-danger-700 focus:ring-danger-500/40 shadow-lg hover:shadow-xl"
  };

  // Size classes using design tokens
  const sizeClasses = {
    xs: "h-(--button-height-xs) px-2 py-1 text-xs gap-1",
    sm: "h-(--button-height-sm) px-3 py-1.5 text-sm gap-1.5",
    md: "h-(--button-height-md) px-4 py-2 text-sm gap-2",
    lg: "h-(--button-height-lg) px-6 py-3 text-base gap-3",
    xl: "h-(--button-height-xl) px-8 py-4 text-lg gap-4"
  };

  // Computed values with $derived
  const interactionClasses = $derived(ripple ? "ripple btn-hover-lift" : "btn-hover-lift");
  const widthClass = $derived(fullWidth ? "w-full" : "");
  const classes = $derived(`${baseClasses} ${variantClasses[variant!]} ${sizeClasses[size!]} ${interactionClasses} ${widthClass} ${className}`.trim());

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
      <LoadingSpinner size="sm" variant="secondary" />
    {:else}
      {@render children?.()}
    {/if}
  </a>
{:else}
  <button {type} {disabled} class={classes} onclick={handleClick}>
    {#if loading}
      <LoadingSpinner size="sm" variant="secondary" />
    {:else}
      {@render children?.()}
    {/if}
  </button>
{/if}

<style>
  /* Ripple effect */
  .ripple {
    position: relative;
    overflow: hidden;
    transform: translateZ(0);
  }

  .ripple::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 0;
    height: 0;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.3);
    transform: translate(-50%, -50%);
    transition: width 0.6s, height 0.6s;
  }

  .ripple:active::before {
    width: 300px;
    height: 300px;
  }

  /* Button hover lift effect */
  .btn-hover-lift {
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .btn-hover-lift:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .btn-hover-lift:active {
    transform: translateY(0);
    transition-duration: 0.1s;
  }
</style>