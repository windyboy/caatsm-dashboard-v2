<!--
  @component LoadingSpinner
  Standardized loading spinner component with consistent API following design system.

  @param {ComponentSize} size - Spinner size
  @param {ComponentVariant} variant - Spinner variant
  @param {string} message - Optional loading message
  @param {boolean} overlay - Whether to show as overlay
  @param {string} overlayColor - Overlay background color
-->

<script lang="ts">
  import type { LoadingSpinnerProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends LoadingSpinnerProps {
    variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'success' | 'warning' | 'danger' | 'gradient';
  }

  let {
    size = 'md',
    variant = 'primary',
    message = '',
    overlay = false,
    overlayColor = 'rgba(255, 255, 255, 0.8)',
    class: className = ''
  }: Props = $props();

  // Size classes using design tokens
  const sizeClasses = {
    xs: 'w-3 h-3',
    sm: 'w-4 h-4',
    md: 'w-8 h-8',
    lg: 'w-12 h-12',
    xl: 'w-16 h-16'
  };

  // Variant classes using design tokens
  const variantClasses = {
    primary: 'border-brand-500',
    secondary: 'border-slate-400',
    outline: 'border-brand-600',
    ghost: 'border-slate-300',
    success: 'border-success-500',
    warning: 'border-warning-500',
    danger: 'border-danger-500',
    gradient: '' // Gradient uses a different rendering approach
  };
</script>

<div class="loading-spinner-wrapper {className}" class:overlay>
  {#if overlay}
    <div class="loading-overlay" style="background-color: {overlayColor}">
      <div class="loading-content">
        <div class="loading-spinner relative {sizeClasses[size!]}">
          <!-- Outer ring -->
          <div class="absolute inset-0 rounded-full border-2 border-slate-200"></div>

          {#if variant === 'gradient'}
            <!-- Gradient spinning ring using conic-gradient -->
            <div class="absolute inset-0 rounded-full animate-spin gradient-spinner"></div>

            <!-- Inner pulse for gradient variant -->
            <div class="absolute inset-1 rounded-full bg-gradient-to-r from-brand-400 via-accent-400 to-success-400 animate-pulse opacity-30"></div>
          {:else}
            <!-- Standard spinning ring with solid border color -->
            <div class="absolute inset-0 rounded-full border-2 border-t-transparent {variantClasses[variant!]} animate-spin"></div>
          {/if}
        </div>

        {#if message}
          <p class="loading-message">{message}</p>
        {/if}
      </div>
    </div>
  {:else}
    <div class="loading-inline">
      <div class="loading-spinner relative {sizeClasses[size!]}">
        <!-- Outer ring -->
        <div class="absolute inset-0 rounded-full border-2 border-slate-200"></div>

        {#if variant === 'gradient'}
          <!-- Gradient spinning ring using conic-gradient -->
          <div class="absolute inset-0 rounded-full animate-spin gradient-spinner"></div>

          <!-- Inner pulse for gradient variant -->
          <div class="absolute inset-1 rounded-full bg-gradient-to-r from-brand-400 via-accent-400 to-success-400 animate-pulse opacity-30"></div>
        {:else}
          <!-- Standard spinning ring with solid border color -->
          <div class="absolute inset-0 rounded-full border-2 border-t-transparent {variantClasses[variant!]} animate-spin"></div>
        {/if}
      </div>

      {#if message}
        <p class="loading-message">{message}</p>
      {/if}
    </div>
  {/if}
</div>

<style>
  .loading-spinner-wrapper {
    display: contents;
  }

  .loading-spinner-wrapper.overlay {
    position: relative;
  }

  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 10;
    display: flex;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(2px);
    -webkit-backdrop-filter: blur(2px);
    transition: opacity 0.2s ease;
  }

  .loading-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--spacing-md);
  }

  .loading-inline {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-lg);
  }

  .loading-message {
    color: var(--color-slate-600);
    font-size: var(--text-sm);
    font-weight: var(--font-medium);
    text-align: center;
    margin: 0;
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  .animate-spin {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .gradient-spinner {
    background: conic-gradient(
      from 0deg,
      transparent 0deg,
      transparent 90deg,
      #3b82f6 90deg,
      #8b5cf6 180deg,
      #10b981 270deg,
      transparent 270deg
    );
    mask: radial-gradient(circle, transparent 0%, transparent calc(50% - 2px), black calc(50% - 2px), black 50%, transparent 50%);
    -webkit-mask: radial-gradient(circle, transparent 0%, transparent calc(50% - 2px), black calc(50% - 2px), black 50%, transparent 50%);
  }
</style>