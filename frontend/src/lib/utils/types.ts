// Type definitions for the application

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

export interface TrafficSummary {
  total: number;
  byPriority: Record<number, number>;
  byType: Record<string, number>;
}

export interface WebSocketMessage {
  type: "message" | "stats-total" | "stats-priority" | "stats-type";
  data: any;
}

export interface SearchFilter {
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

