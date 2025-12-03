<script lang="ts">
  import type { Snippet } from "svelte";
  import { onMount, onDestroy } from "svelte";
  import { createLogger } from "$lib/utils/logger";
  import Button from "./ui/Button.svelte";
  import { isBrowser, getWindow } from "$lib/utils/browser";

  const logger = createLogger("ErrorBoundary");

  // Module-level flag to track if global handlers are registered (singleton pattern)
  // Global handlers are registered once and shared across all instances
  let globalHandlersRegistered = false;

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
    // Only register global handlers once (singleton pattern)
    // Each instance creates its own handler functions, but only the first instance registers them
    const win = getWindow();
    if (win && !globalHandlersRegistered) {
      // Create handler functions for this instance
      errorHandler = (event: ErrorEvent) => {
        // Only handle if we don't already have an error in this instance
        if (!error) {
          event.preventDefault();
          handleError(event.error || new Error(event.message), {
            filename: event.filename,
            lineno: event.lineno,
            colno: event.colno,
          });
        }
      };

      rejectionHandler = (event: PromiseRejectionEvent) => {
        // Only handle if we don't already have an error in this instance
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

      win.addEventListener("error", errorHandler);
      win.addEventListener("unhandledrejection", rejectionHandler);
      globalHandlersRegistered = true;
    }
  });

  onDestroy(() => {
    // Only remove listeners if this instance registered them
    const win = getWindow();
    if (win && globalHandlersRegistered && errorHandler && rejectionHandler) {
      win.removeEventListener("error", errorHandler);
      win.removeEventListener("unhandledrejection", rejectionHandler);
      errorHandler = null;
      rejectionHandler = null;
      globalHandlersRegistered = false;
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
