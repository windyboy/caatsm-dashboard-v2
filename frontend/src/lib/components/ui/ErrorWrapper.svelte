<!--
  @component ErrorWrapper
  Higher-order component that provides error state management for any content.

  @param {string | Error} error - Error to display
  @param {boolean} showError - Whether to show error state
  @param {string} errorTitle - Error title
  @param {string} errorMessage - Error message
  @param {boolean} retryable - Whether to show retry button
  @param {string} retryText - Retry button text
-->

<script lang="ts">
  import type { Snippet } from "svelte";
  import Button from "./Button.svelte";
  import type { WithErrorProps } from "../../types/ui";

  // Props with Svelte 5 $props() rune
  interface Props extends WithErrorProps {
    errorTitle?: string;
    errorMessage?: string;
    retryable?: boolean;
    retryText?: string;
    class?: string;
    onretry?: (event: Event) => void;
    children?: Snippet;
  }

  let {
    error,
    showError = true,
    errorTitle = 'Something went wrong',
    errorMessage = 'An error occurred while loading this content.',
    retryable = false,
    retryText = 'Try Again',
    class: className = '',
    onretry,
    children
  }: Props = $props();

  // Computed error state with $derived
  const hasError = $derived(showError && (error || errorMessage !== 'An error occurred while loading this content.'));
  const displayMessage = $derived(error instanceof Error ? error.message : (error as string) || errorMessage);

  function handleRetry() {
    onretry?.(new Event('retry'));
  }
</script>

<div class="error-wrapper {className}">
  {#if hasError}
    <!-- Error state -->
    <div class="error-state">
      <div class="error-icon">⚠️</div>
      <h3 class="error-title">{errorTitle}</h3>
      <p class="error-message">{displayMessage}</p>

      {#if retryable}
        <Button variant="primary" onclick={handleRetry}>
          {retryText}
        </Button>
      {/if}

      <!-- Development error details -->
      {#if import.meta.env.DEV && error instanceof Error && error.stack}
        <details class="error-details">
          <summary>Error Details (Development)</summary>
          <pre class="error-stack">{error.stack}</pre>
        </details>
      {/if}
    </div>
  {:else}
    <!-- Normal content -->
    {@render children?.()}
  {/if}
</div>

<style>
  .error-wrapper {
    display: contents;
  }

  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: var(--spacing-xl);
    min-height: 200px;
    gap: var(--spacing-md);
  }

  .error-icon {
    font-size: 3rem;
    margin-bottom: var(--spacing-sm);
  }

  .error-title {
    font-size: var(--text-lg);
    font-weight: var(--font-semibold);
    color: var(--color-slate-900);
    margin: 0 0 var(--spacing-sm) 0;
  }

  .error-message {
    color: var(--color-slate-600);
    font-size: var(--text-sm);
    line-height: var(--leading-sm);
    margin: 0 0 var(--spacing-lg) 0;
    max-width: 400px;
  }

  .error-details {
    margin-top: var(--spacing-lg);
    text-align: left;
    width: 100%;
    max-width: 600px;
  }

  .error-details summary {
    cursor: pointer;
    font-weight: var(--font-medium);
    color: var(--color-slate-700);
    margin-bottom: var(--spacing-sm);
  }

  .error-stack {
    background-color: var(--color-slate-50);
    border: 1px solid var(--color-slate-200);
    border-radius: var(--radius-md);
    padding: var(--spacing-md);
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: var(--text-xs);
    color: var(--color-slate-800);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0;
    max-height: 200px;
    overflow-y: auto;
  }
</style>