// REST API client for Go backend

import { createLogger } from "../utils/logger.ts";
import type { SearchResult } from "../utils/types";

const logger = createLogger("API");

function resolveApiBase(): string {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL;
  }

  // Default to backend port when running without the Vite proxy (e.g., preview/SSG)
  if (import.meta.env.DEV) {
    return "http://localhost:3002";
  }

  if (typeof window !== "undefined" && window.location?.origin) {
    return window.location.origin;
  }

  return "http://localhost:3002";
}

const API_BASE_URL = resolveApiBase();

function buildUrl(endpoint: string): string {
  try {
    const url = new URL(endpoint, API_BASE_URL);
    return url.toString();
  } catch (error) {
    logger.error("Failed to build API URL", error, {
      endpoint,
      apiBase: API_BASE_URL,
    });
    return `${API_BASE_URL}${endpoint}`;
  }
}

export interface SearchParams {
  query?: string;
  type?: string;
  source?: string;
  destination?: string;
  priority?: number;
  start_time?: string;
  end_time?: string;
  limit?: number;
  offset?: number;
  sort_by?: string;
  order?: string;
}

export interface StatsTotal {
  total: number;
}

export interface StatsPriority {
  byPriority: Record<number, number>;
}

export interface StatsType {
  byType: Record<string, number>;
}

export interface AutocompleteSuggestion {
  value: string;
  type: string;
  label: string;
}

export interface AutocompleteResult {
  suggestions: AutocompleteSuggestion[] | string[];
}

function buildSearchParamsFromParams(params: SearchParams): URLSearchParams {
  const searchParams = new URLSearchParams();

  if (params.query) searchParams.set("query", params.query);
  if (params.type) searchParams.set("type", params.type);
  if (params.source) searchParams.set("source", params.source);
  if (params.destination) searchParams.set("destination", params.destination);
  if (params.priority !== undefined) searchParams.set("priority", params.priority.toString());
  if (params.start_time) searchParams.set("start_time", params.start_time);
  if (params.end_time) searchParams.set("end_time", params.end_time);
  if (params.limit !== undefined) searchParams.set("limit", params.limit.toString());
  if (params.offset !== undefined) searchParams.set("offset", params.offset.toString());
  if (params.sort_by) searchParams.set("sort_by", params.sort_by);
  if (params.order) searchParams.set("order", params.order);

  return searchParams;
}

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = buildUrl(endpoint);
  const method = options?.method || "GET";

  logger.debug("API request", {
    url,
    method,
  });

  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    logger.error("API response error", new Error(error.error || response.statusText), {
      url,
      method,
      status: response.status,
      statusText: response.statusText,
    });
    throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`);
  }

  logger.info("API response success", {
    url,
    method,
    status: response.status,
  });

  return response.json();
}

export async function search(params: SearchParams): Promise<SearchResult> {
  const searchParams = buildSearchParamsFromParams(params);

  return request<SearchResult>(`/api/search?${searchParams.toString()}`);
}

export async function autocomplete(term: string, size: number = 5): Promise<AutocompleteResult> {
  const params = new URLSearchParams({ term, size: size.toString() });
  return request<AutocompleteResult>(`/api/autocomplete?${params.toString()}`);
}

export async function getStatsTotal(): Promise<StatsTotal> {
  return request<StatsTotal>("/api/stats/total");
}

export async function getStatsPriority(): Promise<StatsPriority> {
  return request<StatsPriority>("/api/stats/priority");
}

export async function getStatsType(): Promise<StatsType> {
  return request<StatsType>("/api/stats/type");
}

export async function exportData(
  format: "csv" | "xlsx" | "pdf",
  params: SearchParams
): Promise<Blob> {
  const searchParams = buildSearchParamsFromParams(params);

  searchParams.set("format", format);

  const url = buildUrl(`/api/export?${searchParams.toString()}`);
  logger.info("Starting export", {
    url,
    format,
  });

  const response = await fetch(url);

  if (!response.ok) {
    logger.error("Export failed", new Error(response.statusText), {
      url,
      status: response.status,
      statusText: response.statusText,
    });
    throw new Error(`Export failed: ${response.statusText}`);
  }

  logger.info("Export success", {
    url,
    status: response.status,
  });

  return response.blob();
}
