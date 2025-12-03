// Type definitions for the application
// These types align with the backend domain types in internal/domain/

/**
 * Telegram represents a domain entity for aviation telegrams.
 * Aligns with backend domain.Telegram struct.
 * Note: time is serialized as ISO string from backend time.Time
 * Note: raw_data is optional and may not be present in API responses (backend uses json:"-")
 */
export interface Telegram {
  message_id: string; // Required, non-empty, no control characters
  type: string; // Required, one of: AFTN, SITA, ACARS, CPDLC
  time: string; // ISO 8601 timestamp string (from backend time.Time)
  flight_number: string; // Uppercase, trimmed
  source: string; // ICAO code (4 uppercase letters), optional
  destination: string; // ICAO code (4 uppercase letters), optional
  priority: number; // 1-3 (1=urgent, 2=operational, 3=routine)
  content: string; // Message content, trimmed
  raw_data?: string; // Optional, may not be present in API responses
}

/**
 * SearchResult wraps the result of a search query.
 * Aligns with backend domain.SearchResult struct.
 * Note: total is int64 in backend, represented as number in JS
 */
export interface SearchResult {
  telegrams: Telegram[];
  total: number; // int64 in backend, number in JS
  page: {
    limit: number; // Non-negative
    offset: number; // Non-negative
    sort_by: string; // One of: time, priority, message_id, type, flight_number, source, destination
    order: string; // "asc" or "desc"
  };
}

/**
 * TrafficSummary contains aggregated traffic metrics.
 * Aligns with backend domain.TrafficSummary struct.
 * Note: All counts are int64 in backend, represented as number in JS
 */
export interface TrafficSummary {
  total: number; // int64 in backend
  byPriority: Record<number, number>; // Priority (1-3) -> count (int64 in backend)
  byType: Record<string, number>; // Type (AFTN/SITA/ACARS/CPDLC) -> count (int64 in backend)
}

// Discriminated union for type-safe WebSocket messages
export type WebSocketMessage =
  | { type: "message"; data: Telegram }
  | { type: "stats"; data: { total: number; byPriority: Record<number, number>; byType: Record<string, number> } }
  | { type: "pong"; data?: never };

/**
 * SearchFilter represents search criteria for telegrams.
 * Note: Backend domain.SearchFilters uses arrays for type/source/destination/priority,
 * but the API accepts single values which are converted to arrays on the backend.
 * This interface matches the API contract (single values).
 * 
 * Time range validation:
 * - start_time must be before end_time
 * - Range cannot exceed 90 days (MaxTimeRangeDays)
 * 
 * Sort fields: time, priority, message_id, type, flight_number, source, destination
 * Order: "asc" or "desc"
 */
export interface SearchFilter {
  query?: string; // Full-text search query
  type?: string; // Single type (AFTN, SITA, ACARS, CPDLC) - backend converts to array
  source?: string; // Single ICAO code - backend converts to array
  destination?: string; // Single ICAO code - backend converts to array
  priority?: number; // Single priority (1-3) - backend converts to array
  start_time?: string; // ISO 8601 timestamp string
  end_time?: string; // ISO 8601 timestamp string
  limit?: number; // Non-negative, default 50
  offset?: number; // Non-negative, default 0
  sort_by?: string; // One of: time, priority, message_id, type, flight_number, source, destination
  order?: string; // "asc" or "desc", default "desc"
}
