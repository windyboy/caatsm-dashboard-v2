import type { HealthSnapshot, HistoricalStats, SearchResponse, TrafficSummary } from "./types";
import { getConnectionStore } from "./stores/connection.svelte";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:3002";

function buildUrl(path: string): string {
  return new URL(path, API_BASE).toString();
}

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

      const response = await fetch(url, {
        ...init,
        signal: init?.signal,
        headers: {
          "Content-Type": "application/json",
          ...init?.headers,
        },
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
      ? `无法连接到服务器。请检查后端服务是否正在运行。`
      : lastError?.message || "请求失败";

  connectionStore.markHttpError(errorMessage);
  throw lastError;
}

export function fetchStats(): Promise<TrafficSummary> {
  // Default to last 24 hours
  const endTime = new Date();
  const startTime = new Date(endTime.getTime() - 24 * 60 * 60 * 1000);

  const params = new URLSearchParams({
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
  });

  return request<TrafficSummary>(`/api/stats?${params.toString()}`);
}

export function fetchTrendData(interval: string = "hour"): Promise<HistoricalStats> {
  // Default to last 24 hours
  const endTime = new Date();
  const startTime = new Date(endTime.getTime() - 24 * 60 * 60 * 1000);

  const params = new URLSearchParams({
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString(),
    interval: interval,
  });

  return request<HistoricalStats>(`/api/stats/historical?${params.toString()}`);
}

export async function fetchHealth(): Promise<HealthSnapshot> {
  // Health endpoint may return 503 for degraded/unhealthy status, which is still valid data
  // We need to handle 503 as a valid response since it contains health information
  const connectionStore = getConnectionStore();
  const url = buildUrl("/api/health");
  const maxRetries = 3;
  let lastError: Error | null = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      connectionStore.markHttpSuccess();

      const response = await fetch(url, {
        headers: { "Content-Type": "application/json" },
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
      ? `无法连接到服务器。请检查后端服务是否正在运行。`
      : lastError?.message || "请求失败";

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
