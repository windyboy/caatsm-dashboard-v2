import type { HealthSnapshot, HistoricalStats, SearchResponse, TrafficSummary } from "./types";
import { getConnectionStore } from "./stores/connection.svelte";
import { authManager } from "./auth";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:3002";

function buildUrl(path: string): string {
  return new URL(path, API_BASE).toString();
}

/**
 * Makes an HTTP request with retry logic and error handling.
 * @param path - API endpoint path
 * @param init - Request options including signal for abort and retry flag
 * @returns Promise resolving to the response data
 */
async function request<T>(
  path: string,
  init?: RequestInit & { signal?: AbortSignal; retry?: boolean }
): Promise<T> {
  const connectionStore = getConnectionStore();
  const url = buildUrl(path);
  const shouldRetry = init?.retry !== false;
  const maxRetries = 3;
  let lastError: Error | null = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      connectionStore.markHttpSuccess();

      // 构建请求头
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };

      // 合并现有头
      if (init?.headers) {
        if (init.headers instanceof Headers) {
          init.headers.forEach((value, key) => {
            headers[key] = value;
          });
        } else {
          Object.assign(headers, init.headers);
        }
      }

      // 添加认证头
      const accessToken = authManager.getAccessToken();
      if (accessToken) {
        headers["Authorization"] = `Bearer ${accessToken}`;
      }

      // 添加CSRF token（对于非GET请求）
      if (init?.method && init.method !== "GET") {
        const csrfToken = authManager.getCsrfToken();
        if (csrfToken) {
          headers["X-CSRF-Token"] = csrfToken;
        }
      }

      const response = await fetch(url, {
        ...init,
        signal: init?.signal,
        headers,
      });

      if (!response.ok) {
        let errorMessage = response.statusText || "Request failed";
        try {
          const errorBody = await response.json();
          if (errorBody.error || errorBody.message) {
            errorMessage = errorBody.error || errorBody.message;
          }
        } catch {
          // Use status text if JSON parsing fails
        }
        throw new Error(`${response.status}: ${errorMessage}`);
      }

      if (response.status === 204) {
        return {} as T;
      }

      return response.json() as Promise<T>;
    } catch (err) {
      lastError = err instanceof Error ? err : new Error(String(err));

      // Don't retry on abort or if retry is disabled
      if (err instanceof Error && err.name === "AbortError") {
        throw err;
      }
      if (!shouldRetry || attempt === maxRetries) {
        break;
      }

      // Network errors - retry with exponential backoff
      if (err instanceof TypeError && err.message.includes("fetch")) {
        const delay = Math.min(1000 * Math.pow(2, attempt), 5000);
        await new Promise((resolve) => setTimeout(resolve, delay));
        continue;
      }

      // HTTP errors - only retry on 5xx or network issues
      if (lastError.message.includes("Network error") || lastError.message.match(/^5\d{2}:/)) {
        const delay = Math.min(1000 * Math.pow(2, attempt), 5000);
        await new Promise((resolve) => setTimeout(resolve, delay));
        continue;
      }

      // Don't retry other errors
      break;
    }
  }

  // All retries failed
  const errorMessage =
    lastError instanceof TypeError && lastError.message.includes("fetch")
      ? "Unable to connect to server. Please check if the backend service is running."
      : lastError?.message || "Request failed";

  connectionStore.markHttpError(errorMessage);
  throw lastError;
}

/**
 * Calculates time range based on time window string
 * @param timeWindow - Time window string: "last_1h", "last_24h", "last_7d", or "custom"
 * @returns Object with startTime and endTime
 */
function calculateTimeRange(timeWindow: string): { startTime: Date; endTime: Date } {
  const endTime = new Date();
  let startTime: Date;

  switch (timeWindow) {
    case "last_1h":
      startTime = new Date(endTime.getTime() - 60 * 60 * 1000);
      break;
    case "last_24h":
      startTime = new Date(endTime.getTime() - 24 * 60 * 60 * 1000);
      break;
    case "last_7d":
      startTime = new Date(endTime.getTime() - 7 * 24 * 60 * 60 * 1000);
      break;
    default:
      // Default to last 24 hours
      startTime = new Date(endTime.getTime() - 24 * 60 * 60 * 1000);
  }

  return { startTime, endTime };
}

export function fetchStats(timeWindow: string = "last_24h"): Promise<TrafficSummary> {
  const { startTime, endTime } = calculateTimeRange(timeWindow);

  const params = new URLSearchParams({
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
  });

  return request<TrafficSummary>(`/api/stats?${params.toString()}`);
}

export function fetchTrendData(
  timeWindow: string = "last_24h",
  interval: string = "hour"
): Promise<HistoricalStats> {
  const { startTime, endTime } = calculateTimeRange(timeWindow);

  // Adjust interval based on time window
  let adjustedInterval = interval;
  if (timeWindow === "last_7d") {
    adjustedInterval = "day";
  } else if (timeWindow === "last_1h") {
    adjustedInterval = "hour";
  }

  const params = new URLSearchParams({
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
    interval: adjustedInterval,
  });

  return request<HistoricalStats>(`/api/stats/historical?${params.toString()}`);
}

export async function fetchHealth(): Promise<HealthSnapshot> {
  // Health endpoint may return 503 for degraded/unhealthy status, which is still valid data
  // We need to handle 503 as a valid response since it contains health information
  // Use the shared request function but handle 503 as a valid status
  const connectionStore = getConnectionStore();
  const url = buildUrl("/api/health");
  const maxRetries = 3;
  let lastError: Error | null = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      connectionStore.markHttpSuccess();

      // Build headers with authentication
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };

      const accessToken = authManager.getAccessToken();
      if (accessToken) {
        headers["Authorization"] = `Bearer ${accessToken}`;
      }

      const response = await fetch(url, {
        headers,
      });

      // 503 is valid for health endpoint - it means degraded/unhealthy but still has data
      if (response.status === 503 || response.status === 200) {
        return await response.json();
      }
      // For other error statuses, throw an error
      if (!response.ok) {
        let errorMessage = response.statusText || "Request failed";
        try {
          const errorBody = await response.json();
          if (errorBody.error || errorBody.message) {
            errorMessage = errorBody.error || errorBody.message;
          }
        } catch {
          // Use status text if JSON parsing fails
        }
        throw new Error(`${response.status}: ${errorMessage}`);
      }
      return await response.json();
    } catch (err) {
      lastError = err instanceof Error ? err : new Error(String(err));

      if (attempt === maxRetries) {
        break;
      }

      // Don't retry on abort
      if (err instanceof Error && err.name === "AbortError") {
        throw err;
      }

      // Network errors - retry with exponential backoff
      if (err instanceof TypeError && err.message.includes("fetch")) {
        const delay = Math.min(1000 * Math.pow(2, attempt), 5000);
        await new Promise((resolve) => setTimeout(resolve, delay));
        continue;
      }

      // HTTP errors - only retry on 5xx or network issues
      if (lastError.message.includes("Network error") || lastError.message.match(/^5\d{2}:/)) {
        const delay = Math.min(1000 * Math.pow(2, attempt), 5000);
        await new Promise((resolve) => setTimeout(resolve, delay));
        continue;
      }

      // Don't retry other errors
      break;
    }
  }

  // All retries failed
  const errorMessage =
    lastError instanceof TypeError && lastError.message.includes("fetch")
      ? "Unable to connect to server. Please check if the backend service is running."
      : lastError?.message || "Request failed";

  connectionStore.markHttpError(errorMessage);
  throw lastError;
}

export function fetchRecentMessages(limit: number = 12): Promise<SearchResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    sort_by: "time",
    order: "desc",
  });

  return request<SearchResponse>(`/api/search?${params.toString()}`);
}

/**
 * Converts datetime-local format to RFC3339 format with timezone offset.
 * @param dateTimeLocal - Date string in datetime-local format (YYYY-MM-DDTHH:mm)
 * @returns RFC3339 formatted date string
 */
function formatDateTimeLocalToRFC3339(dateTimeLocal: string): string {
  if (!dateTimeLocal) return "";
  // datetime-local format: "YYYY-MM-DDTHH:mm"
  // Parse as local time and convert to RFC3339 with timezone offset
  // Split the datetime-local string into date and time parts
  const [datePart, timePart] = dateTimeLocal.split("T");
  if (!datePart || !timePart) {
    // Fallback: try Date constructor
    const date = new Date(dateTimeLocal);
    if (isNaN(date.getTime())) {
      return `${dateTimeLocal}:00Z`;
    }
    return date.toISOString();
  }

  // Create a date object in local timezone
  // datetime-local values are always in local time
  const [year, month, day] = datePart.split("-").map(Number);
  const [hours, minutes] = timePart.split(":").map(Number);

  // Create date in local timezone
  const localDate = new Date(year, month - 1, day, hours, minutes);

  // Convert to ISO string (RFC3339 format with timezone offset)
  return localDate.toISOString();
}

export function runSearch(
  query: string,
  limit: number = 50,
  signal?: AbortSignal,
  startTime?: string,
  endTime?: string
): Promise<SearchResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    sort_by: "time",
    order: "desc",
  });

  if (query.trim()) {
    params.set("query", query.trim());
  }

  if (startTime) {
    params.set("start_time", formatDateTimeLocalToRFC3339(startTime));
  }

  if (endTime) {
    params.set("end_time", formatDateTimeLocalToRFC3339(endTime));
  }

  return request<SearchResponse>(`/api/search?${params.toString()}`, { signal });
}
