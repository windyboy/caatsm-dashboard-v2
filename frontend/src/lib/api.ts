import type { HealthSnapshot, SearchResponse, TrafficSummary } from "./types";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:3002";

function buildUrl(path: string): string {
  return new URL(path, API_BASE).toString();
}

async function request<T>(path: string, init?: RequestInit & { signal?: AbortSignal }): Promise<T> {
  const url = buildUrl(path);
  
  try {
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
    if (err instanceof Error && err.name === "AbortError") {
      throw err;
    }
    if (err instanceof TypeError && err.message.includes("fetch")) {
      throw new Error(`Network error: Unable to connect to ${url}. Is the backend server running?`);
    }
    throw err;
  }
}

export function fetchStats(): Promise<TrafficSummary> {
  return request<TrafficSummary>("/api/stats");
}

export function fetchHealth(): Promise<HealthSnapshot> {
  // Health endpoint may return 503 for degraded/unhealthy status, which is still valid data
  // We need to handle 503 as a valid response since it contains health information
  const url = buildUrl("/api/health");
  return fetch(url, {
    headers: { "Content-Type": "application/json" },
  }).then(async (response) => {
    // 503 is valid for health endpoint - it means degraded/unhealthy but still has data
    if (response.status === 503 || response.status === 200) {
      return response.json() as Promise<HealthSnapshot>;
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
    return response.json() as Promise<HealthSnapshot>;
  }).catch((err) => {
    if (err instanceof TypeError && err.message.includes("fetch")) {
      throw new Error(`Network error: Unable to connect to ${url}. Is the backend server running?`);
    }
    throw err;
  });
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
  // RFC3339 format: "YYYY-MM-DDTHH:mm:ssZ"
  return `${dateTimeLocal}:00Z`;
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
