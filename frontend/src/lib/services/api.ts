// REST API client for Go backend

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:3002";

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

export interface SearchResult {
  telegrams: Telegram[];
  total: number;
  page: {
    limit: number;
    offset: number;
    sort_by: string;
    order: string;
  };
}

export interface Telegram {
  message_id: string;
  type: string;
  time: string;
  flight_number: string;
  source: string;
  destination: string;
  priority: number;
  content: string;
  raw_data?: string;
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

export interface AutocompleteResult {
  suggestions: string[];
}

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`);
  }

  return response.json();
}

export async function search(params: SearchParams): Promise<SearchResult> {
  const searchParams = new URLSearchParams();
  
  if (params.query) searchParams.set("query", params.query);
  if (params.type) searchParams.set("type", params.type);
  if (params.source) searchParams.set("source", params.source);
  if (params.destination) searchParams.set("destination", params.destination);
  if (params.priority !== undefined) searchParams.set("priority", params.priority.toString());
  if (params.start_time) searchParams.set("start_time", params.start_time);
  if (params.end_time) searchParams.set("end_time", params.end_time);
  if (params.limit) searchParams.set("limit", params.limit.toString());
  if (params.offset) searchParams.set("offset", params.offset.toString());
  if (params.sort_by) searchParams.set("sort_by", params.sort_by);
  if (params.order) searchParams.set("order", params.order);

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
  const searchParams = new URLSearchParams();
  
  if (params.query) searchParams.set("query", params.query);
  if (params.type) searchParams.set("type", params.type);
  if (params.source) searchParams.set("source", params.source);
  if (params.destination) searchParams.set("destination", params.destination);
  if (params.priority !== undefined) searchParams.set("priority", params.priority.toString());
  searchParams.set("format", format);

  const url = `${API_BASE_URL}/api/export?${searchParams.toString()}`;
  const response = await fetch(url);

  if (!response.ok) {
    throw new Error(`Export failed: ${response.statusText}`);
  }

  return response.blob();
}

