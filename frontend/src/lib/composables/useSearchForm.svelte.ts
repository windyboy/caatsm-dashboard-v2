/**
 * Composable for search form state and validation
 */

import { SEARCH_CONFIG } from "$lib/constants";
import type { SearchParams } from "$lib/services/api";

export function useSearchForm(
  initialParams?:
    | {
        query?: string;
        type?: string;
        priority?: string;
        start_time?: string;
        end_time?: string;
      }
    | (() => {
        query?: string;
        type?: string;
        priority?: string;
        start_time?: string;
        end_time?: string;
      } | undefined)
) {
  const getParams = typeof initialParams === "function" ? initialParams : () => initialParams;

  let query = $state(getParams()?.query || "");
  let type = $state(getParams()?.type || "");
  let priority = $state(getParams()?.priority || "");
  let start_time = $state(getParams()?.start_time || "");
  let end_time = $state(getParams()?.end_time || "");
  let error = $state<string | null>(null);

  // Sync state when initialParams changes
  $effect(() => {
    const params = getParams();
    if (params) {
      query = params.query || "";
      type = params.type || "";
      priority = params.priority || "";
      start_time = params.start_time || "";
      end_time = params.end_time || "";
    }
  });

  // Convert datetime-local (local time) to ISO string (UTC) - API expects UTC timestamps
  function toIsoString(value: string): string | undefined {
    if (!value) {
      return undefined;
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return undefined;
    }
    return date.toISOString();
  }

  function validate(): boolean {
    error = null;

    const startISO = toIsoString(start_time);
    const endISO = toIsoString(end_time);

    // Validate time range if both dates are provided
    if (startISO && endISO) {
      const start = new Date(startISO);
      const end = new Date(endISO);

      // Check if start time is before end time
      if (start >= end) {
        error = "Start time must be before end time";
        return false;
      }

      // Check if time range exceeds maximum (90 days)
      const diffMs = end.getTime() - start.getTime();
      const diffDays = diffMs / (1000 * 60 * 60 * 24);

      if (diffDays > SEARCH_CONFIG.MAX_TIME_RANGE_DAYS) {
        error = `Time range cannot exceed ${SEARCH_CONFIG.MAX_TIME_RANGE_DAYS} days`;
        return false;
      }
    }

    return true;
  }

  function getSearchParams(): SearchParams {
    const startISO = toIsoString(start_time);
    const endISO = toIsoString(end_time);

    return {
      query,
      type: type || undefined,
      priority: priority ? parseInt(priority) : undefined,
      start_time: startISO,
      end_time: endISO,
    };
  }

  function reset() {
    query = "";
    type = "";
    priority = "";
    start_time = "";
    end_time = "";
    error = null;
  }

  return {
    get query() {
      return query;
    },
    set query(value: string) {
      query = value;
    },
    get type() {
      return type;
    },
    set type(value: string) {
      type = value;
    },
    get priority() {
      return priority;
    },
    set priority(value: string) {
      priority = value;
    },
    get start_time() {
      return start_time;
    },
    set start_time(value: string) {
      start_time = value;
    },
    get end_time() {
      return end_time;
    },
    set end_time(value: string) {
      end_time = value;
    },
    get error() {
      return error;
    },
    validate,
    getSearchParams,
    reset,
  };
}

