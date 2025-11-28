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
    ripple = true,
    type = "button",
    href = undefined,
    fullWidth = false,
    class: className = "",
    onclick,
    children,
  }: Props = $props();

  // Base classes using Daisy UI
  const baseClasses = "btn";

  // Variant classes using Daisy UI
  const variantClasses = {
    primary: "btn-primary",
    secondary: "btn-secondary",
    outline: "btn-outline",
    ghost: "btn-ghost",
    success: "btn-success",
    warning: "btn-warning",
    danger: "btn-error",
  };

  // Size classes using Daisy UI
  const sizeClasses = {
    xs: "btn-xs",
    sm: "btn-sm",
    md: "btn-md",
    lg: "btn-lg",
    xl: "btn-xl",
  };

  // Computed values with $derived
  const loadingClass = $derived(loading ? "loading" : "");
  const widthClass = $derived(fullWidth ? "w-full" : "");
  const classes = $derived(
    `${baseClasses} ${variantClasses[variant!]} ${sizeClasses[size!]} ${loadingClass} ${widthClass} ${className}`.trim()
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
    {@render children?.()}
  </button>
{/if}
