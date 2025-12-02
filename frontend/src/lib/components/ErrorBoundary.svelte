<script lang="ts">
  import type { Snippet } from "svelte";
  import { onMount, onDestroy } from "svelte";
  import { createLogger } from "$lib/utils/logger";
  import Button from "./ui/Button.svelte";

  const logger = createLogger("ErrorBoundary");

  interface Props {
    fallbackMessage?: string;
    showRetry?: boolean;
    context?: Record<string, unknown>;
    onerror?: (detail: { error: Error; context?: Record<string, unknown> }) => void;
    children?: Snippet<
      [{ handleError: (err: Error | unknown, additionalContext?: Record<string, unknown>) => void }]
    >;
  }

  let {
    fallbackMessage = "Something went wrong. Please try again.",
    showRetry = true,
    context = {},
    onerror,
    children,
  }: Props = $props();

  let error = $state<Error | null>(null);
  let errorId = $state<string | null>(null);
  let errorHandler: ((event: ErrorEvent) => void) | null = null;
  let rejectionHandler: ((event: PromiseRejectionEvent) => void) | null = null;

  export function handleError(
    err: Error | unknown,
    additionalContext: Record<string, unknown> = {}
  ) {
    const errorObj = err instanceof Error ? err : new Error(String(err));
    error = errorObj;
    errorId = `error-${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;

    const fullContext = { ...context, ...additionalContext, errorId };

    logger.error("Component error caught by ErrorBoundary", errorObj, fullContext);
    onerror?.({ error: errorObj, context: fullContext });
  }

  function retry() {
    error = null;
    errorId = null;
  }

  onMount(() => {
    // Catch unhandled errors in this component tree
    errorHandler = (event: ErrorEvent) => {
      // Only handle if we don't already have an error
      if (!error) {
        event.preventDefault();
        handleError(event.error || new Error(event.message), {
          filename: event.filename,
          lineno: event.lineno,
          colno: event.colno,
        });
      }
    };

    // Catch unhandled Promise rejections
    rejectionHandler = (event: PromiseRejectionEvent) => {
      // Only handle if we don't already have an error
      if (!error) {
        event.preventDefault();
        handleError(event.reason || new Error("Unhandled promise rejection"), {
          type: "promise_rejection",
          reason: event.reason instanceof Error
            ? {
                name: event.reason.name,
                message: event.reason.message,
                stack: event.reason.stack,
              }
            : event.reason,
        });
      }
    };

    if (typeof window !== "undefined") {
      window.addEventListener("error", errorHandler);
      window.addEventListener("unhandledrejection", rejectionHandler);
    }
  });

  onDestroy(() => {
    if (typeof window !== "undefined") {
      if (errorHandler) {
        window.removeEventListener("error", errorHandler);
        errorHandler = null;
      }
      if (rejectionHandler) {
        window.removeEventListener("unhandledrejection", rejectionHandler);
        rejectionHandler = null;
      }
    }
  });
</script>

{#if error}
  <div class="error-boundary" data-error-id={errorId}>
    <div class="error-content">
      <div class="error-icon">⚠️</div>
      <h3 class="error-title">Oops! Something went wrong</h3>
      <p class="error-message">{fallbackMessage}</p>
      {#if import.meta.env.DEV}
        <details class="error-details">
          <summary>Error Details (Development)</summary>
          <pre class="error-stack">{error.message}</pre>
          {#if error.stack}
            <pre class="error-stack">{error.stack}</pre>
          {/if}
        </details>
      {/if}
      {#if showRetry}
        <Button onclick={retry} variant="primary">Try Again</Button>
      {/if}
    </div>
  </div>
{:else}
  {@render children?.({ handleError })}
{/if}

<style>
  .error-boundary {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 200px;
    padding: 2rem;
    border: 1px solid var(--color-border);
    border-radius: 8px;
    background-color: var(--color-surface);
  }

  .error-content {
    text-align: center;
    max-width: 400px;
  }

  .error-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
  }

  .error-title {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--color-text);
    margin-bottom: 0.5rem;
  }

  .error-message {
    color: var(--color-text-secondary);
    margin-bottom: 1.5rem;
    line-height: 1.5;
  }

  .error-details {
    margin-bottom: 1.5rem;
    text-align: left;
  }

  .error-details summary {
    cursor: pointer;
    font-weight: 500;
    margin-bottom: 0.5rem;
  }

  .error-stack {
    background-color: var(--color-code-bg);
    border: 1px solid var(--color-border);
    border-radius: 4px;
    padding: 1rem;
    font-family: monospace;
    font-size: 0.875rem;
    color: var(--color-text);
    white-space: pre-wrap;
    word-break: break-word;
    margin: 0.5rem 0;
    max-height: 200px;
    overflow-y: auto;
  }
</style>
