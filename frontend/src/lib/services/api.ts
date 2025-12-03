// REST API client for Go backend

import { createLogger } from "../utils/logger";
import type { SearchResult } from "../utils/types";
import { API_CONFIG } from "../constants";

const logger = createLogger("API");

function getApiBaseUrl(): string {
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

function buildUrl(endpoint: string): string {
  const apiBaseUrl = getApiBaseUrl();
  try {
    const url = new URL(endpoint, apiBaseUrl);
    return url.toString();
  } catch (error) {
    logger.error("Failed to build API URL", error, {
      endpoint,
      apiBase: apiBaseUrl,
    });

    // Normalize base: trim trailing slashes
    const normalizedBase = apiBaseUrl.replace(/\/+$/, "");

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

export interface TrafficSummary {
  total: number;
  byPriority: Record<number, number>;
  byType: Record<string, number>;
}

// Legacy interfaces for backward compatibility (if needed)
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

export interface ComponentHealth {
  status: string; // "ok", "error", "not_configured"
  message?: string;
}

export interface HealthCheckResult {
  status: string; // "ok" or "degraded"
  timestamp: string;
  checks: {
    postgresql: ComponentHealth;
    meilisearch: ComponentHealth;
    redis: ComponentHealth;
    nats: ComponentHealth;
  };
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

export async function request<T>(
  endpoint: string,
  options?: RequestInit & { signal?: AbortSignal; retries?: number }
): Promise<T> {
  const url = buildUrl(endpoint);
  const method = options?.method || "GET";
  const maxRetries = options?.retries ?? API_CONFIG.MAX_RETRIES;
  let lastError: Error | null = null;

  // Create abort controller for timeout (30 seconds)
  const timeoutController = new AbortController();
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  // Combine signals if both exist
  const createCombinedSignal = (): AbortSignal => {
    if (!options?.signal) {
      return timeoutController.signal;
    }
    const combined = new AbortController();
    options.signal.addEventListener('abort', () => combined.abort());
    timeoutController.signal.addEventListener('abort', () => combined.abort());
    return combined.signal;
  };

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      // Set timeout for this attempt (30 seconds)
      timeoutId = setTimeout(() => {
        timeoutController.abort();
      }, 30000);

      const combinedSignal = createCombinedSignal();

      logger.debug("API request", { url, method, attempt });

      const response = await fetch(url, {
        ...options,
        signal: combinedSignal,
        headers: {
          "Content-Type": "application/json",
          ...options?.headers,
        },
      });

      // Clear timeout on successful response
      if (timeoutId) {
        clearTimeout(timeoutId);
        timeoutId = null;
      }

      if (!response.ok) {
        // Try to parse error response, but preserve original error information
        let errorData: { error?: string; [key: string]: unknown } = { error: response.statusText };
        try {
          const contentType = response.headers.get("content-type");
          if (contentType && contentType.includes("application/json")) {
            errorData = await response.json();
          }
        } catch (parseError) {
          // If JSON parsing fails, log the parse error but keep the statusText
          logger.warn("Failed to parse error response as JSON", {
            url,
            status: response.status,
            contentType: response.headers.get("content-type"),
            parseError: parseError instanceof Error
              ? {
                  name: parseError.name,
                  message: parseError.message,
                  stack: parseError.stack,
                }
              : parseError,
          });
        }
        
        // Preserve original error information while using parsed data if available
        const errorMessage = errorData.error || `HTTP ${response.status}: ${response.statusText}`;
        lastError = new Error(errorMessage);
        
        // Attach additional error context if available
        if (errorData && Object.keys(errorData).length > 1) {
          (lastError as Error & { context?: Record<string, unknown> }).context = errorData;
        }

        // Handle 401 Unauthorized - redirect to login
        if (response.status === 401) {
          logger.warn("Authentication required", {
            url,
            method,
          });
          
          if (typeof window !== "undefined") {
            // Store current URL for redirect after login
            sessionStorage.setItem("redirectAfterLogin", window.location.pathname + window.location.search);
            // Redirect to login page
            window.location.href = "/login";
          }
          
          throw lastError;
        }

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
      // Clear timeout on error
      if (timeoutId) {
        clearTimeout(timeoutId);
        timeoutId = null;
      }

      lastError = error instanceof Error ? error : new Error(String(error));

      // Don't retry if request was aborted (timeout or manual abort)
      if (lastError.name === "AbortError") {
        if (timeoutController.signal.aborted) {
          lastError = new Error("Request timeout: The server did not respond in time");
        }
        logger.error("API request aborted", lastError, {
          url,
          method,
          attempt,
        });
        throw lastError;
      }

      // Don't retry TypeError (CORS, network, etc.) beyond max attempts
      if (attempt >= maxRetries) {
        logger.error("API request failed (max retries)", lastError, {
          url,
          method,
          attempts: maxRetries + 1,
        });
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

  // Clean up timeout if still set
  if (timeoutId) {
    clearTimeout(timeoutId);
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

/**
 * Get telegram statistics from the server.
 * Returns total count, breakdown by priority, and breakdown by type.
 * Defaults to last 24 hours if no time range is provided.
 */
export async function getStats(startTime?: string, endTime?: string): Promise<TrafficSummary> {
  const params = new URLSearchParams();
  if (startTime) {
    params.set("start_time", startTime);
  }
  if (endTime) {
    params.set("end_time", endTime);
  }
  
  // Default to last 24 hours if no time range provided
  if (!startTime && !endTime) {
    const end = new Date();
    const start = new Date(end.getTime() - 24 * 60 * 60 * 1000); // 24 hours ago
    params.set("start_time", start.toISOString());
    params.set("end_time", end.toISOString());
  }
  
  const queryString = params.toString();
  return request<TrafficSummary>(`/api/stats${queryString ? `?${queryString}` : ""}`);
}

// Legacy functions for backward compatibility (if needed)
export async function getStatsTotal(): Promise<StatsTotal> {
  return request<StatsTotal>("/api/stats/total");
}

export async function getStatsPriority(): Promise<StatsPriority> {
  return request<StatsPriority>("/api/stats/priority");
}

export async function getStatsType(): Promise<StatsType> {
  return request<StatsType>("/api/stats/type");
}

/**
 * Get system health status from the server.
 * Returns health status of all components (PostgreSQL, Redis, Meilisearch, NATS).
 */
export async function getHealthStatus(): Promise<HealthCheckResult> {
  return request<HealthCheckResult>("/api/health");
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
