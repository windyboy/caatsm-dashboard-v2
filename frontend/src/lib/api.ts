import type { HealthSnapshot, SearchResponse, TrafficSummary } from "./types";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:3002";

function buildUrl(path: string): string {
  return new URL(path, API_BASE).toString();
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const url = buildUrl(path);
  
  try {
    const response = await fetch(url, {
      ...init,
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
        // If JSON parsing fails, use status text
      }
      throw new Error(`${response.status}: ${errorMessage}`);
    }

    if (response.status === 204) {
      return {} as T;
    }

    return response.json() as Promise<T>;
  } catch (err) {
    // Handle network errors (connection refused, CORS, etc.)
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
  return request<HealthSnapshot>("/api/health");
}

export function fetchRecentMessages(limit: number = 12): Promise<SearchResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    sort_by: "time",
    order: "desc",
  });

  return request<SearchResponse>(`/api/search?${params.toString()}`);
}

export function runSearch(query: string, limit: number = 50): Promise<SearchResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    sort_by: "time",
    order: "desc",
  });

  if (query.trim()) {
    params.set("query", query.trim());
  }

  return request<SearchResponse>(`/api/search?${params.toString()}`);
}
