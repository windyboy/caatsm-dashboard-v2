import { onDestroy } from "svelte";
import { runSearch } from "$lib/api";
import type { Telegram } from "$lib/types";

export interface UseSearchOptions {
  getQuery: () => string;
  getStartTime: () => string;
  getEndTime: () => string;
}

export function useSearch(options: UseSearchOptions) {
  const { getQuery, getStartTime, getEndTime } = options;

  let results = $state<Telegram[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let abortController: AbortController | null = null;

  const hasSearchCriteria = $derived.by(() => {
    const query = getQuery();
    const startTime = getStartTime();
    const endTime = getEndTime();
    return query.trim().length > 0 || startTime.length > 0 || endTime.length > 0;
  });

  async function performSearch() {
    if (abortController) {
      abortController.abort();
    }
    abortController = new AbortController();
    loading = true;
    error = null;

    try {
      const query = getQuery();
      const startTime = getStartTime();
      const endTime = getEndTime();

      const response = await runSearch(
        query,
        1000,
        abortController.signal,
        startTime || undefined,
        endTime || undefined
      );
      if (abortController.signal.aborted) return;
      results = response.telegrams ?? [];
      total = response.total ?? results.length;
    } catch (err) {
      if (err instanceof Error && err.name === "AbortError") return;
      results = [];
      total = 0;
      error = err instanceof Error ? err.message : "Search failed.";
    } finally {
      if (!abortController.signal.aborted) {
        loading = false;
      }
    }
  }

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();

    if (!hasSearchCriteria) {
      error = "Please enter a search query or time range";
      return;
    }

    performSearch();
  }

  function handleValidationError(validationError: string) {
    error = validationError;
    results = [];
    total = 0;
  }

  function handleReset() {
    error = null;
    results = [];
    total = 0;
  }

  onDestroy(() => {
    if (abortController) abortController.abort();
  });

  return {
    results,
    total,
    loading,
    error,
    hasSearchCriteria,
    performSearch,
    handleSubmit,
    handleValidationError,
    handleReset,
  };
}
