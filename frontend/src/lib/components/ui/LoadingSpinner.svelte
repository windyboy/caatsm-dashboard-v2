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

  // Size classes using Daisy UI
  const sizeClasses = {
    xs: "loading-xs",
    sm: "loading-sm",
    md: "loading-md",
    lg: "loading-lg",
    xl: "loading-xl",
  };

  // Variant classes using Daisy UI
  const variantClasses = {
    primary: "text-primary",
    secondary: "text-secondary",
    outline: "text-primary",
    ghost: "text-base-content",
    success: "text-success",
    warning: "text-warning",
    danger: "text-error",
    gradient: "text-primary", // or appropriate gradient styling
  };
</script>

<div class="loading-spinner-wrapper {className}" class:overlay>
  {#if overlay}
    <div class="loading-overlay" style="background-color: {overlayColor}">
      <div class="loading-content">
        <span class="loading loading-spinner {sizeClasses[size!]} {variantClasses[variant!]}"
        ></span>

        {#if message}
          <p class="loading-message">{message}</p>
        {/if}
      </div>
    </div>
  {:else}
    <div class="loading-inline">
      <span class="loading loading-spinner {sizeClasses[size!]} {variantClasses[variant!]}"></span>

      {#if message}
        <p class="loading-message">{message}</p>
      {/if}
    </div>
  {/if}
</div>
