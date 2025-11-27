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

  // Base classes using design tokens
  const baseClasses = "rounded-lg backdrop-blur-md border-0 transition-all duration-300";

  // Variant classes using design tokens
  const variantClasses = {
    default: "bg-white/95 card-glow",
    elevated: "bg-white/98 shadow-xl card-glow",
    bordered: "bg-white/90 border-2 border-brand-200/50",
    filled: "bg-brand-50/80 border border-brand-100"
  };

  // Size-based padding using design tokens
  const sizePadding = {
    xs: "p-(--card-padding-xs)",
    sm: "p-(--card-padding-sm)",
    md: "p-(--card-padding-md)",
    lg: "p-(--card-padding-lg)",
    xl: "p-(--card-padding-xl)"
  };

  // Computed values with $derived
  const hoverClasses = $derived(hover ? "card-hover-pulse" : "");
  const interactiveClasses = $derived(interactive ? "cursor-pointer focus:outline-none focus:ring-2 focus:ring-brand-500/40" : "");
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

<style>
  /* Card glow effect */
  .card-glow {
    background:
      linear-gradient(135deg,
        rgba(255, 255, 255, 0.95) 0%,
        rgba(252, 252, 253, 0.9) 50%,
        rgba(248, 250, 252, 0.95) 100%
      );
    backdrop-filter: blur(20px) saturate(180%);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
    box-shadow:
      0 8px 32px rgba(59, 130, 246, 0.08),
      0 4px 16px rgba(0, 0, 0, 0.04),
      0 2px 8px rgba(0, 0, 0, 0.02),
      0 0 0 1px rgba(226, 232, 240, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.8);
    border: 1px solid rgba(226, 232, 240, 0.6);
    position: relative;
    transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .card-glow::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg,
      transparent 0%,
      rgba(59, 130, 246, 0.2) 20%,
      rgba(168, 85, 247, 0.2) 50%,
      rgba(34, 197, 94, 0.2) 80%,
      transparent 100%
    );
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  .card-glow:hover {
    background:
      linear-gradient(135deg,
        rgba(255, 255, 255, 0.98) 0%,
        rgba(248, 250, 252, 0.95) 50%,
        rgba(241, 245, 249, 0.98) 100%
      );
    backdrop-filter: blur(24px) saturate(200%);
    -webkit-backdrop-filter: blur(24px) saturate(200%);
    box-shadow:
      0 20px 64px rgba(59, 130, 246, 0.12),
      0 8px 32px rgba(0, 0, 0, 0.06),
      0 4px 16px rgba(0, 0, 0, 0.04),
      0 0 0 1px rgba(203, 213, 225, 0.5),
      inset 0 1px 0 rgba(255, 255, 255, 1);
    border-color: rgba(203, 213, 225, 0.7);
    transform: translateY(-2px) scale(1.01);
  }

  .card-glow:hover::before {
    opacity: 1;
  }

  /* Card hover pulse effect */
  .card-hover-pulse {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .card-hover-pulse:hover {
    animation: cardPulse 2s ease-in-out infinite;
  }

  @keyframes cardPulse {
    0%, 100% {
      transform: scale(1);
    }
    50% {
      transform: scale(1.005);
    }
  }
</style>