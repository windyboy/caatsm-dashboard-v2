<!--
  @component Card
  Standardized card wrapper component with consistent API following design system.

  @param {CardProps['variant']} variant - Card style variant
  @param {ComponentSize} size - Card size affecting padding
  @param {boolean} hover - Whether to enable hover effects
  @param {boolean} padding - Whether to apply default padding
  @param {boolean} interactive - Whether the card is interactive/clickable
  @param {string} className - Additional CSS classes
-->

<script lang="ts">
  import type { Snippet } from "svelte";
  import type { CardProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends CardProps {
    onclick?: (event: Event) => void;
    children?: Snippet;
  }

  let {
    variant = 'default',
    size = 'md',
    hover = true,
    padding = true,
    interactive = false,
    class: className = '',
    onclick,
    children
  }: Props = $props();

  // Base classes using Daisy UI
  const baseClasses = "card";

  // Variant classes using Daisy UI
  const variantClasses = {
    default: "bg-base-100",
    elevated: "bg-base-100 shadow-xl",
    bordered: "bg-base-100 border border-base-300",
    filled: "bg-primary/10 border border-primary/20"
  };

  // Size-based padding using Daisy UI
  const sizePadding = {
    xs: "card-body p-2",
    sm: "card-body p-4",
    md: "card-body p-6",
    lg: "card-body p-8",
    xl: "card-body p-10"
  };

  // Computed values with $derived
  const hoverClasses = $derived(hover ? "hover:shadow-lg transition-shadow" : "");
  const interactiveClasses = $derived(interactive ? "cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary" : "");
  const paddingClasses = $derived(padding ? sizePadding[size!] : "");
  const classes = $derived(`${baseClasses} ${variantClasses[variant!]} ${paddingClasses} ${hoverClasses} ${interactiveClasses} ${className}`.trim());

  function handleClick(event: Event) {
    if (interactive) {
      onclick?.(event);
    }
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (interactive && (event.key === 'Enter' || event.key === ' ')) {
      event.preventDefault();
      onclick?.(event);
    }
  }
</script>

{#if interactive}
  <button
    type="button"
    class={classes}
    onclick={handleClick}
    onkeydown={handleKeyDown}
  >
    {@render children?.()}
  </button>
{:else}
  <div class={classes}>
    {@render children?.()}
  </div>
{/if}
