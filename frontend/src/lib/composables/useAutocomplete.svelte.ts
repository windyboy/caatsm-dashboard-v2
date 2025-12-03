/**
 * Composable for autocomplete functionality with race condition protection
 */

import { autocomplete } from "$lib/services/api";
import { createLogger } from "$lib/utils/logger";
import { SEARCH_CONFIG } from "$lib/constants";

const logger = createLogger("useAutocomplete");

export function useAutocomplete() {
  let suggestions = $state<Array<{ value: string; type: string; label: string } | string>>([]);
  let showSuggestions = $state(false);
  let autocompleteTimeout = $state<ReturnType<typeof setTimeout> | null>(null);
  let abortController = $state<AbortController | null>(null);
  let autocompleteRequestId = $state(0);

  function cleanup() {
    if (autocompleteTimeout !== null) {
      clearTimeout(autocompleteTimeout);
      autocompleteTimeout = null;
    }
    if (abortController) {
      abortController.abort();
      abortController = null;
    }
  }

  async function handleAutocomplete(query: string) {
    // Cancel previous request
    if (abortController) {
      abortController.abort();
    }

    if (autocompleteTimeout) {
      clearTimeout(autocompleteTimeout);
    }

    if (query.length < SEARCH_CONFIG.AUTOCOMPLETE_MIN_LENGTH) {
      suggestions = [];
      showSuggestions = false;
      return;
    }

    // Increment request ID to track the "latest" request
    const requestId = ++autocompleteRequestId;
    const currentQuery = query; // Capture at debounce start

    autocompleteTimeout = setTimeout(async () => {
      // Double-check we're still the latest request
      if (requestId !== autocompleteRequestId) {
        logger.debug("Autocomplete request superseded", { requestId, latest: autocompleteRequestId });
        return;
      }

      abortController = new AbortController();
      const signal = abortController.signal;

      try {
        const result = await autocomplete(currentQuery, 5, signal);

        // Triple-check: query unchanged, not aborted, still latest request
        if (
          requestId === autocompleteRequestId &&
          currentQuery === query &&
          !signal.aborted
        ) {
          // Handle both old format (string[]) and new format (AutocompleteSuggestion[])
          if (result.suggestions.length > 0 && typeof result.suggestions[0] === "string") {
            // Old format: convert to new format
            suggestions = (result.suggestions as string[]).map((s) => ({
              value: s,
              type: "text",
              label: "",
            }));
          } else {
            suggestions = result.suggestions as Array<{
              value: string;
              type: string;
              label: string;
            }>;
          }
          showSuggestions = suggestions.length > 0;
        }
      } catch (error) {
        // Ignore AbortError - it's expected when request is cancelled
        if (error instanceof Error && error.name !== "AbortError") {
          logger.error("Autocomplete failed", error, {
            query: currentQuery,
            queryLength: currentQuery.length,
            requestId,
          });
        }
        // Only clear suggestions if request wasn't aborted and is still latest
        if (requestId === autocompleteRequestId && !signal.aborted) {
          suggestions = [];
          showSuggestions = false;
        }
      }
    }, SEARCH_CONFIG.AUTOCOMPLETE_DEBOUNCE_MS);
  }

  function clearSuggestions() {
    suggestions = [];
    showSuggestions = false;
  }

  // Cleanup on component destroy
  $effect(() => {
    return () => {
      cleanup();
    };
  });

  return {
    get suggestions() {
      return suggestions;
    },
    get showSuggestions() {
      return showSuggestions;
    },
    handleAutocomplete,
    clearSuggestions,
    cleanup,
  };
}

