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

  // Variant classes using Daisy UI
  const variantClasses = {
    primary: "alert-info",
    secondary: "alert",
    outline: "alert-info",
    ghost: "alert",
    success: "alert-success",
    warning: "alert-warning",
    danger: "alert-error"
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
        clearTimeoutFn();
        // Don't call ondismiss here - ontimeout already indicates dismissal
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

  <div
    class="toast-container fixed z-50 {positionClasses[position!]}"
    transition:slide={{ duration: 300, axis: 'y' }}
    onmouseenter={handleMouseEnter}
    onmouseleave={handleMouseLeave}
  >
  >
    <div class="alert {variantClasses[variant!]} {className}" role="alert">
      <!-- Icon -->
      {#if currentIcon}
        <span>{currentIcon}</span>
      {/if}

      <!-- Content -->
      <div>
        {#if title}
          <div class="font-bold">{title}</div>
        {/if}
        {#if message}
          <div>{message}</div>
        {/if}
      </div>

      <!-- Close button -->
      {#if dismissible}
        <button
          type="button"
          class="btn btn-sm btn-circle btn-ghost"
          onclick={dismiss}
          aria-label="Dismiss toast"
        >
          ✕
        </button>
      {/if}
    </div>
  </div>
{/if}
