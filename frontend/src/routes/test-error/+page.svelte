<script lang="ts">
  import ErrorBoundary from '$lib/components/ErrorBoundary.svelte';

  // This route is for E2E testing error boundaries
  // It intentionally triggers an error in the error boundary
</script>

{#if import.meta.env.DEV || import.meta.env.MODE === 'test'}
  <ErrorBoundary 
    fallbackMessage="This is a test error boundary. The component encountered an error."
    showRetry={true}
    context={{ testRoute: true }}
  >
    {#snippet children({ handleError })}
      <div data-testid="error-trigger-page" class="test-error-page">
        <h1>Error Boundary Test Page</h1>
        <p>This page is only available in development and test environments.</p>
        <button 
          data-testid="trigger-error-button"
          onclick={() => {
            const testError = new Error('Test error for visual regression testing');
            handleError(testError);
          }}
          class="trigger-button"
        >
          Trigger Error
        </button>
      </div>
    {/snippet}
  </ErrorBoundary>
{:else}
  <div>
    <h1>404 - Page Not Found</h1>
    <p>This test page is only available in development mode.</p>
  </div>
{/if}

<style>
  .test-error-page {
    padding: 2rem;
    max-width: 600px;
    margin: 0 auto;
  }

  h1 {
    color: var(--color-text);
    margin-bottom: 1rem;
  }

  p {
    color: var(--color-text-secondary);
    margin-bottom: 2rem;
    line-height: 1.6;
  }

  .trigger-button {
    background-color: #dc2626;
    color: white;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 6px;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .trigger-button:hover {
    background-color: #b91c1c;
  }

  .trigger-button:active {
    background-color: #991b1b;
  }
</style>

