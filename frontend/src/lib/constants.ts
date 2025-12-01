// Application constants

// Search configuration
export const SEARCH_CONFIG = {
  MAX_TIME_RANGE_DAYS: 90,
  AUTOCOMPLETE_DEBOUNCE_MS: 500,
  AUTOCOMPLETE_MIN_LENGTH: 2,
} as const;

// WebSocket configuration
export const WEBSOCKET_CONFIG = {
  MAX_MESSAGE_SIZE: 1048576, // 1MB
  MAX_HANDLERS: 100,
  PING_INTERVAL_MS: 30000, // 30 seconds
  PONG_TIMEOUT_MS: 5000, // 5 seconds
  INITIAL_RECONNECT_DELAY_MS: 1000, // 1 second
  MAX_RECONNECT_DELAY_MS: 30000, // 30 seconds
  MAX_RECONNECT_ATTEMPTS: 10,
} as const;

// API configuration
export const API_CONFIG = {
  MAX_RETRIES: 3,
  RETRY_MAX_DELAY_MS: 5000, // 5 seconds
} as const;

// UI configuration
export const UI_CONFIG = {
  SLOW_REQUEST_TIMEOUT_MS: 3000, // 3 seconds
  MAX_MESSAGES: 50,
} as const;

