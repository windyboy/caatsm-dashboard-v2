// REST API client for Go backend

import { createLogger } from "../utils/logger";
import type { SearchResult } from "../utils/types";
import { API_CONFIG } from "../constants";

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

    // Normalize base: trim trailing slashes
    const normalizedBase = API_BASE_URL.replace(/\/+$/, "");

    // Normalize endpoint: ensure it starts with exactly one slash
    let normalizedEndpoint = endpoint.trim();
    if (normalizedEndpoint === "") {
      normalizedEndpoint = "/";
    } else {
      // Remove any leading slashes and add exactly one
      normalizedEndpoint = "/" + normalizedEndpoint.replace(/^\/+/, "");
    }

    return normalizedBase + normalizedEndpoint;
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

async function request<T>(
  endpoint: string,
  options?: RequestInit & { signal?: AbortSignal; retries?: number }
): Promise<T> {
  const url = buildUrl(endpoint);
  const method = options?.method || "GET";
  const maxRetries = options?.retries ?? API_CONFIG.MAX_RETRIES;
  let lastError: Error | null = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      logger.debug("API request", { url, method, attempt });

      const response = await fetch(url, {
        ...options,
        signal: options?.signal,
        headers: {
          "Content-Type": "application/json",
          ...options?.headers,
        },
      });

      if (!response.ok) {
        const error = await response.json().catch(() => ({ error: response.statusText }));
        const errorMessage = error.error || `HTTP ${response.status}: ${response.statusText}`;
        lastError = new Error(errorMessage);

        // Don't retry client errors (4xx except 429)
        if (response.status >= 400 && response.status < 500 && response.status !== 429) {
          logger.error("API client error (non-retryable)", lastError, {
            url,
            method,
            status: response.status,
          });
          throw lastError;
        }

        // Retry server errors and rate limits
        if (attempt < maxRetries) {
          const delay = Math.min(
            1000 * Math.pow(2, attempt),
            API_CONFIG.RETRY_MAX_DELAY_MS
          ); // Exponential backoff, max 5s
          logger.warn("API error, retrying", {
            url,
            method,
            status: response.status,
            attempt,
            delayMs: delay,
          });
          await new Promise((resolve) => setTimeout(resolve, delay));
          continue;
        }

        logger.error("API response error (max retries)", lastError, {
          url,
          method,
          status: response.status,
          attempts: maxRetries + 1,
        });
        throw lastError;
      }

      logger.info("API response success", { url, method, status: response.status, attempt });
      return response.json();
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error));

      // Don't retry TypeError (CORS, network, etc.) beyond max attempts
      if (attempt >= maxRetries) {
        logger.error("API request failed (max retries)", lastError, {
          url,
          method,
          attempts: maxRetries + 1,
        });
        throw lastError;
      }

      // Don't retry if request was aborted
      if (lastError.name === "AbortError") {
        throw lastError;
      }

      const delay = Math.min(1000 * Math.pow(2, attempt), API_CONFIG.RETRY_MAX_DELAY_MS);
      logger.warn("API request error, retrying", {
        url,
        method,
        error: lastError.message,
        attempt,
        delayMs: delay,
      });
      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  }

  throw lastError || new Error("Request failed");
}

export async function search(params: SearchParams): Promise<SearchResult> {
  const searchParams = buildSearchParamsFromParams(params);

  return request<SearchResult>(`/api/search?${searchParams.toString()}`);
}

export async function autocomplete(
  term: string,
  size: number = 5,
  signal?: AbortSignal
): Promise<AutocompleteResult> {
  const params = new URLSearchParams({ term, size: size.toString() });
  return request<AutocompleteResult>(`/api/autocomplete?${params.toString()}`, { signal });
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
