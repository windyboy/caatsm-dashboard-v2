export interface TrafficSummary {
  total: number;
  byType?: Record<string, number>;
  activeRoutes?: number;
  messagesPerSec?: number;
  timeWindow?: string;
}

export interface ComponentHealth {
  status?: string;
  message?: string;
  response_time?: string;
}

export interface HealthSnapshot {
  status?: string;
  version?: string;
  timestamp?: string;
  service?: string;
  postgresql?: ComponentHealth;
  meilisearch?: ComponentHealth;
  redis?: ComponentHealth;
  nats?: ComponentHealth;
}

export interface Telegram {
  message_id?: string;
  type?: string;
  time?: string;
  flight_number?: string;
  source?: string;
  destination?: string;
  priority?: number;
  content?: string;
  raw_data?: string;
}

export interface SearchResponse {
  telegrams: Telegram[];
  total: number;
}

// WebSocket message types
export interface WSMessage {
  type: string;
  data: unknown;
}

export interface WSStatsData {
  total: number;
  byType?: Record<string, number>;
  activeRoutes?: number;
  messagesPerSec?: number;
  timeWindow?: string;
}

export const WSMessageType = {
  STATS: "stats",
  MESSAGE: "message",
} as const;

export type WSConnectionStatus = "connecting" | "connected" | "disconnected" | "error";
