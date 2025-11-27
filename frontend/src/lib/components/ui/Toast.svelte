<!--
  @component Toast
  Standardized toast/notification component with consistent API following design system.

  @param {ComponentVariant} variant - Toast variant
  @param {string} title - Toast title
  @param {string} message - Toast message
  @param {number} duration - Toast duration in milliseconds
  @param {boolean} dismissible - Whether toast can be dismissed
  @param {string} icon - Custom icon
  @param {'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'top-center' | 'bottom-center'} position - Toast position
-->

<script lang="ts">
  import { fade, slide } from "svelte/transition";
  import type { ToastProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends ToastProps {
    ondismiss?: (event: Event) => void;
    ontimeout?: (event: Event) => void;
  }

  let {
    variant = 'primary',
    title,
    message = '',
    duration = 5000,
    dismissible = true,
    icon,
    position = 'top-right',
    class: className = '',
    ondismiss,
    ontimeout
  }: Props = $props();

  // Internal state
  let mounted = $state(false);
  let timeoutId = $state<number | undefined>(undefined);

  // Position classes
  const positionClasses = {
    'top-left': 'top-4 left-4',
    'top-right': 'top-4 right-4',
    'bottom-left': 'bottom-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'top-center': 'top-4 left-1/2 transform -translate-x-1/2',
    'bottom-center': 'bottom-4 left-1/2 transform -translate-x-1/2'
  };

  // Variant classes using design tokens
  const variantClasses = {
    primary: "bg-white border-brand-200 text-slate-900",
    secondary: "bg-slate-50 border-slate-200 text-slate-700",
    outline: "bg-transparent border-brand-500 text-brand-600",
    ghost: "bg-white/90 border-transparent text-slate-700 backdrop-blur-sm",
    success: "bg-success-50 border-success-200 text-success-800",
    warning: "bg-warning-50 border-warning-200 text-warning-800",
    danger: "bg-danger-50 border-danger-200 text-danger-800"
  };

  // Icon defaults based on variant
  const defaultIcons = {
    primary: 'ℹ️',
    secondary: '💬',
    outline: 'ℹ️',
    ghost: '💭',
    success: '✅',
    warning: '⚠️',
    danger: '❌'
  };

  const currentIcon = $derived(icon || defaultIcons[variant!]);

  function startTimeout() {
    if (duration && duration > 0) {
      timeoutId = window.setTimeout(() => {
        ontimeout?.(new Event('timeout'));
        dismiss();
      }, duration);
    }
  }

  function clearTimeoutFn() {
    if (timeoutId) {
      window.clearTimeout(timeoutId);
      timeoutId = undefined;
    }
  }

  function dismiss() {
    clearTimeoutFn();
    ondismiss?.(new Event('dismiss'));
  }

  function handleMouseEnter() {
    clearTimeoutFn();
  }

  function handleMouseLeave() {
    startTimeout();
  }

  // Lifecycle with $effect
  $effect(() => {
    mounted = true;
    startTimeout();
    
    return () => {
      clearTimeoutFn();
    };
  });
</script>

{#if mounted}
  <div
    class="toast-container {positionClasses[position!]}"
    transition:slide={{ duration: 300, axis: (position || 'top-right').includes('top') ? 'y' : 'y' }}
    onmouseenter={handleMouseEnter}
    onmouseleave={handleMouseLeave}
  >
    <div class="toast-content {variantClasses[variant!]} {className}" role="alert">
      <!-- Icon -->
      {#if currentIcon}
        <div class="toast-icon">
          {currentIcon}
        </div>
      {/if}

      <!-- Content -->
      <div class="toast-body">
        {#if title}
          <div class="toast-title">{title}</div>
        {/if}
        {#if message}
          <div class="toast-message">{message}</div>
        {/if}
      </div>

      <!-- Close button -->
      {#if dismissible}
        <button
          type="button"
          class="toast-close"
          onclick={dismiss}
          aria-label="Dismiss toast"
        >
          ×
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .toast-container {
    position: fixed;
    z-index: 1100;
    pointer-events: none;
  }

  .toast-content {
    pointer-events: auto;
    display: flex;
    align-items: flex-start;
    gap: var(--spacing-sm);
    padding: var(--spacing-md);
    border-radius: var(--radius-lg);
    border: 1px solid;
    box-shadow: var(--shadow-lg);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    max-width: 400px;
    min-width: 300px;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .toast-content:hover {
    transform: translateY(-1px);
    box-shadow: var(--shadow-xl);
  }

  .toast-icon {
    font-size: 1.25rem;
    line-height: 1;
    flex-shrink: 0;
    margin-top: 0.125rem;
  }

  .toast-body {
    flex: 1;
    min-width: 0;
  }

  .toast-title {
    font-size: var(--text-sm);
    font-weight: var(--font-semibold);
    line-height: var(--leading-sm);
    margin-bottom: var(--spacing-xs);
  }

  .toast-message {
    font-size: var(--text-sm);
    line-height: var(--leading-sm);
    color: var(--color-slate-600);
  }

  .toast-close {
    background: none;
    border: none;
    color: var(--color-slate-400);
    cursor: pointer;
    font-size: 1.25rem;
    line-height: 1;
    padding: var(--spacing-xs);
    border-radius: var(--radius-md);
    transition: all 0.2s ease;
    flex-shrink: 0;
    margin-top: -0.25rem;
    margin-right: -0.25rem;
    width: 1.5rem;
    height: 1.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .toast-close:hover {
    background-color: var(--color-slate-100);
    color: var(--color-slate-600);
  }

  /* Variant-specific styles */
  .toast-content.bg-success-50 {
    border-color: var(--color-success-200);
  }

  .toast-content.bg-success-50 .toast-icon {
    color: var(--color-success-600);
  }

  .toast-content.bg-warning-50 {
    border-color: var(--color-warning-200);
  }

  .toast-content.bg-warning-50 .toast-icon {
    color: var(--color-warning-600);
  }

  .toast-content.bg-danger-50 {
    border-color: var(--color-danger-200);
  }

  .toast-content.bg-danger-50 .toast-icon {
    color: var(--color-danger-600);
  }

  /* Animation keyframes for progress bar (future enhancement) */
  @keyframes toast-progress {
    from {
      width: 100%;
    }
    to {
      width: 0%;
    }
  }

  /* Responsive adjustments */
  @media (max-width: 640px) {
    .toast-container {
      left: var(--spacing-sm) !important;
      right: var(--spacing-sm) !important;
      top: auto !important;
      bottom: var(--spacing-sm) !important;
      transform: none !important;
    }

    .toast-content {
      max-width: none;
      min-width: auto;
    }
  }
</style>