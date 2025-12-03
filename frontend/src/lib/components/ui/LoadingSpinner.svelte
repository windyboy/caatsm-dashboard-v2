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
  interface Props extends LoadingSpinnerProps {}

  let {
    size = "md",
    variant = "primary",
    message = "",
    overlay = false,
    overlayColor = "rgba(255, 255, 255, 0.8)",
    class: className = "",
  }: Props = $props();

  // Size classes using UnoCSS
  const sizeClasses = {
    xs: "w-3 h-3",
    sm: "w-4 h-4",
    md: "w-6 h-6",
    lg: "w-8 h-8",
    xl: "w-10 h-10",
  };

  // Variant classes using UnoCSS
  const variantClasses = {
    primary: "text-brand-500",
    secondary: "text-slate-500",
    outline: "text-brand-500",
    ghost: "text-slate-600",
    success: "text-success-500",
    warning: "text-warning-500",
    danger: "text-danger-500",
    gradient: "text-brand-500",
  };
</script>

<div class="loading-spinner-wrapper {className}" class:overlay>
  {#if overlay}
    <div class="fixed inset-0 flex items-center justify-center z-50" style="background-color: {overlayColor}">
      <div class="flex flex-col items-center gap-2">
        <span class="inline-block border-2 border-current border-t-transparent rounded-full animate-spin {sizeClasses[size!]} {variantClasses[variant!]}"></span>
        {#if message}
          <p class="text-sm text-slate-600">{message}</p>
        {/if}
      </div>
    </div>
  {:else}
    <div class="inline-flex items-center gap-2">
      <span class="inline-block border-2 border-current border-t-transparent rounded-full animate-spin {sizeClasses[size!]} {variantClasses[variant!]}"></span>
      {#if message}
        <p class="text-sm text-slate-600">{message}</p>
      {/if}
    </div>
  {/if}
</div>
