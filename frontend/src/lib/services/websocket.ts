// WebSocket client for real-time updates

import { createLogger } from "../utils/logger.ts";

const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:3002/ws";
const logger = createLogger("WebSocket");

// Constants
const INITIAL_RECONNECT_DELAY = 1000; // 1 second
const MAX_RECONNECT_DELAY = 30000; // 30 seconds
const MAX_RECONNECT_ATTEMPTS = 10;
const MAX_MESSAGE_SIZE = 1048576; // 1MB
const MAX_HANDLERS = 100;
const MAX_LOG_PREVIEW_LENGTH = 50; // Limit log preview to prevent PII exposure

export interface WebSocketMessage {
  type: "message" | "stats-total" | "stats-priority" | "stats-type";
  data: any;
}

export type WebSocketMessageHandler = (message: WebSocketMessage) => void;

/**
 * WebSocket client with automatic reconnection and exponential backoff.
 *
 * Features:
 * - Automatic reconnection with exponential backoff (capped at 30 seconds)
 * - Message validation and size limits
 * - Handler subscription management
 * - Safe error handling that doesn't break other handlers
 *
 * @example
 * ```typescript
 * const client = new WebSocketClient();
 * const unsubscribe = client.subscribe((msg) => console.log(msg));
 * client.connect();
 * // Later...
 * unsubscribe();
 * client.disconnect();
 * ```
 */
export class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private readonly maxReconnectAttempts = MAX_RECONNECT_ATTEMPTS;
  private reconnectDelay = INITIAL_RECONNECT_DELAY;
  private readonly maxReconnectDelay = MAX_RECONNECT_DELAY;
  private handlers: Set<WebSocketMessageHandler> = new Set();
  private shouldReconnect = true;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  /**
   * Connects to the WebSocket server. Safe to call multiple times.
   * If already connected, this is a no-op.
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      logger.debug("Already connected, skipping connection attempt");
      return; // Already connected
    }

    try {
      this.ws = new WebSocket(WS_URL);

      this.ws.onopen = () => {
        logger.info("Connection established", {
          url: WS_URL,
          readyState: this.ws?.readyState,
        });
        this.reconnectAttempts = 0;
        this.reconnectDelay = INITIAL_RECONNECT_DELAY;
      };

      this.ws.onmessage = (event) => {
        try {
          // Check message size before parsing
          const dataLength = typeof event.data === "string" ? event.data.length : event.data?.byteLength || 0;
          if (dataLength > MAX_MESSAGE_SIZE) {
            logger.warn("Message exceeds size limit", {
              dataLength,
              maxSize: MAX_MESSAGE_SIZE,
              preview: typeof event.data === "string" ? event.data.substring(0, MAX_LOG_PREVIEW_LENGTH) : undefined,
            });
            return;
          }

          const parsed = JSON.parse(event.data);
          if (!this.validateMessage(parsed)) {
            logger.warn("Invalid message structure", {
              dataLength,
              preview: typeof event.data === "string" ? event.data.substring(0, MAX_LOG_PREVIEW_LENGTH) : undefined,
            });
            return;
          }

          // Ensure data field exists (default to undefined if missing, e.g., when data: undefined was stringified)
          const message: WebSocketMessage = {
            type: (parsed as any).type,
            data: "data" in parsed ? (parsed as any).data : undefined,
          };
          logger.debug("Message received", {
            type: message.type,
            dataSize: message.data !== undefined ? JSON.stringify(message.data).length : 0,
          });
          this.handlers.forEach((handler) => {
            try {
              handler(message);
            } catch (error) {
              logger.error("Handler error", error, {
                messageType: message.type,
                handlerCount: this.handlers.size,
              });
            }
          });
        } catch (error) {
          logger.error("Failed to parse message", error, {
            dataLength: typeof event.data === "string" ? event.data.length : event.data?.byteLength || 0,
            preview: typeof event.data === "string" ? event.data.substring(0, MAX_LOG_PREVIEW_LENGTH) : undefined,
          });
        }
      };

      this.ws.onerror = (event) => {
        logger.error("Connection error", new Error("WebSocket connection error"), {
          eventType: event.type,
          readyState: this.ws?.readyState,
          url: WS_URL,
          reconnectAttempts: this.reconnectAttempts,
        });
      };

      this.ws.onclose = (event) => {
        logger.info("Connection closed", {
          code: event.code,
          reason: event.reason || "No reason provided",
          wasClean: event.wasClean,
          reconnectAttempts: this.reconnectAttempts,
          willReconnect: this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts,
        });
        this.ws = null;

        if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.scheduleReconnect();
        } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          logger.warn("Max reconnection attempts reached", {
            maxAttempts: this.maxReconnectAttempts,
          });
        }
      };
    } catch (error) {
      logger.error("Failed to create connection", error, {
        url: WS_URL,
        reconnectAttempts: this.reconnectAttempts,
      });
      this.scheduleReconnect();
    }
  }

  /**
   * Validates that a parsed message matches the expected structure.
   * Note: data field is optional (undefined values are omitted in JSON.stringify).
   */
  private validateMessage(data: unknown): data is WebSocketMessage {
    if (typeof data !== "object" || data === null) {
      return false;
    }
    const msg = data as Record<string, unknown>;
    if (typeof msg.type !== "string") {
      return false;
    }
    const validTypes = ["message", "stats-total", "stats-priority", "stats-type"];
    if (!validTypes.includes(msg.type)) {
      return false;
    }
    // data field is optional - if missing, it will be set to undefined
    return true;
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer !== null) {
      return; // Already scheduled
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.maxReconnectDelay
    );

    logger.info("Scheduling reconnection", {
      delayMs: delay,
      attempt: this.reconnectAttempts,
      maxAttempts: this.maxReconnectAttempts,
      nextDelay: delay,
    });

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      logger.debug("Executing reconnection attempt", {
        attempt: this.reconnectAttempts,
      });
      this.connect();
    }, delay);
  }

  /**
   * Disconnects from the WebSocket server and prevents automatic reconnection.
   * Safe to call multiple times (idempotent).
   */
  disconnect(): void {
    logger.info("Disconnecting", {
      wasConnected: this.isConnected(),
      hasReconnectTimer: this.reconnectTimer !== null,
    });
    this.shouldReconnect = false;
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Subscribes a handler function to receive WebSocket messages.
   * Returns an unsubscribe function.
   *
   * @param handler - Function to call when messages are received
   * @returns Unsubscribe function
   */
  subscribe(handler: WebSocketMessageHandler): () => void {
    if (this.handlers.size >= MAX_HANDLERS) {
      logger.warn("Handler limit reached", {
        currentCount: this.handlers.size,
        maxHandlers: MAX_HANDLERS,
      });
    }
    this.handlers.add(handler);
    return () => {
      this.handlers.delete(handler);
    };
  }

  /**
   * Checks if the WebSocket is currently connected.
   * @returns true if connected, false otherwise
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // Test helper methods (only used in tests)
  getReconnectAttempts(): number {
    return this.reconnectAttempts;
  }

  getReconnectDelay(): number {
    return this.reconnectDelay;
  }

  getHandlerCount(): number {
    return this.handlers.size;
  }

  getShouldReconnect(): boolean {
    return this.shouldReconnect;
  }
}

// Singleton instance
export const wsClient = new WebSocketClient();

