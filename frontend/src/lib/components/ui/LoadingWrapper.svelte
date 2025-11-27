<!--
  @component LoadingWrapper
  Higher-order component that provides loading state management for any content.

  @param {boolean} loading - Whether to show loading state
  @param {string} loadingText - Text to display during loading
  @param {boolean} overlay - Whether to show overlay during loading
  @param {string} overlayColor - Overlay background color
  @param {ComponentSize} size - Size of loading spinner
  @param {ComponentVariant} variant - Variant of loading spinner
-->

<script lang="ts">
  import type { Snippet } from "svelte";
  import LoadingSpinner from "./LoadingSpinner.svelte";
  import type { WithLoadingProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends WithLoadingProps {
    overlay?: boolean;
    overlayColor?: string;
    size?: 'sm' | 'md' | 'lg' | 'xl';
    variant?: 'primary' | 'secondary' | 'gradient';
    class?: string;
    children?: Snippet;
  }

  let {
    loading = false,
    loadingText = '',
    overlay = false,
    overlayColor = 'rgba(255, 255, 255, 0.8)',
    size = 'md',
    variant = 'primary',
    class: className = '',
    children
  }: Props = $props();
</script>

<div class="loading-wrapper {className}">
  {#if loading && overlay}
    <!-- Overlay loading -->
    <div class="loading-overlay" style="background-color: {overlayColor}">
      <div class="loading-content">
        <LoadingSpinner {size} {variant} message={loadingText} />
      </div>
    </div>
  {/if}

  <!-- Main content -->
  <div class="loading-main-content" class:disabled={loading && overlay}>
    {@render children?.()}
  </div>

  {#if loading && !overlay}
    <!-- Inline loading -->
    <div class="loading-inline">
      <LoadingSpinner {size} {variant} message={loadingText} />
    </div>
  {/if}
</div>

<style>
  .loading-wrapper {
    position: relative;
    display: contents;
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

  .loading-main-content.disabled {
    pointer-events: none;
    user-select: none;
  }

  .loading-inline {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-lg);
  }
</style>