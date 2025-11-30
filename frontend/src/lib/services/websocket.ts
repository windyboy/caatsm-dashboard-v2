// WebSocket client for real-time updates

import { createLogger } from "../utils/logger.ts";
import type { WebSocketMessage } from "../utils/types.ts";

function resolveWebSocketUrl(): string {
  if (import.meta.env.VITE_WS_URL) {
    return import.meta.env.VITE_WS_URL;
  }

  // Prefer backend port directly when running without the Vite proxy
  if (import.meta.env.DEV) {
    return "ws://localhost:3002/ws";
  }

  if (typeof window !== "undefined" && window.location) {
    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    return `${protocol}://${window.location.host}/ws`;
  }

  return "ws://localhost:3002/ws";
}

const logger = createLogger("WebSocket");

// Constants
const INITIAL_RECONNECT_DELAY = 1000; // 1 second
const MAX_RECONNECT_DELAY = 30000; // 30 seconds
const MAX_RECONNECT_ATTEMPTS = 10;
const MAX_MESSAGE_SIZE = 1048576; // 1MB
const MAX_HANDLERS = 100;
const MAX_LOG_PREVIEW_LENGTH = 50; // Limit log preview to prevent PII exposure
const PING_INTERVAL = 30000; // 30 seconds
const PONG_TIMEOUT = 5000; // 5 seconds

export type WebSocketStatus = "connecting" | "connected" | "disconnected" | "error" | "reconnecting";

export type WebSocketMessageHandler = (message: WebSocketMessage) => void;
export type WebSocketStatusHandler = (status: WebSocketStatus) => void;

/**
 * WebSocket client with automatic reconnection and exponential backoff.
 *
 * Features:
 * - Automatic reconnection with exponential backoff and jitter (capped at 30 seconds)
 * - Message validation and size limits
 * - Handler subscription management
 * - Safe error handling that doesn't break other handlers
 * - Visibility API integration (pauses when tab hidden)
 * - Ping/pong health monitoring
 * - Detailed status tracking with event notifications
 *
 * @example
 * ```typescript
 * const client = new WebSocketClient();
 * const unsubscribe = client.subscribe((msg) => console.log(msg));
 * const statusUnsubscribe = client.onStatusChange((status) => console.log(status));
 * client.connect();
 * // Later...
 * unsubscribe();
 * statusUnsubscribe();
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
  private statusHandlers: Set<WebSocketStatusHandler> = new Set();
  private shouldReconnect = true;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pingTimer: ReturnType<typeof setInterval> | null = null;
  private pongTimer: ReturnType<typeof setTimeout> | null = null;
  private currentStatus: WebSocketStatus = "disconnected";
  private visibilityHandler: (() => void) | null = null;
  private isPaused = false;

  constructor() {
    this.setupVisibilityHandling();
  }

  private getWebSocketUrl(): string {
    return resolveWebSocketUrl();
  }

  private setupVisibilityHandling(): void {
    if (typeof window === "undefined") return;

    this.visibilityHandler = () => {
      if (document.hidden) {
        logger.info("Tab hidden, pausing WebSocket");
        this.pause();
      } else {
        logger.info("Tab visible, resuming WebSocket");
        this.resume();
      }
    };

    document.addEventListener("visibilitychange", this.visibilityHandler);
  }

  private setStatus(status: WebSocketStatus): void {
    if (this.currentStatus === status) return;
    this.currentStatus = status;
    this.statusHandlers.forEach((handler) => {
      try {
        handler(status);
      } catch (error) {
        logger.error("Status handler error", error);
      }
    });
  }

  /**
   * Connects to the WebSocket server. Safe to call multiple times.
   * If already connected, this is a no-op.
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      logger.debug("Already connected, skipping connection attempt");
      return; // Already connected
    }

    const wsUrl = this.getWebSocketUrl();
    this.shouldReconnect = true;

    try {
      logger.info("Attempting WebSocket connection", {
        url: wsUrl,
        hasReconnectTimer: this.reconnectTimer !== null,
        reconnectAttempts: this.reconnectAttempts,
        isPaused: this.isPaused,
      });
      this.setStatus("connecting");
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        logger.info("Connection established", {
          url: wsUrl,
          readyState: this.ws?.readyState,
        });
        this.reconnectAttempts = 0;
        this.reconnectDelay = INITIAL_RECONNECT_DELAY;
        this.setStatus("connected");
        this.startPing();
      };

      this.ws.onmessage = (event) => {
        try {
          // Check message size before parsing
          const dataLength =
            typeof event.data === "string" ? event.data.length : event.data?.byteLength || 0;
          if (dataLength > MAX_MESSAGE_SIZE) {
            logger.warn("Message exceeds size limit", {
              dataLength,
              maxSize: MAX_MESSAGE_SIZE,
              preview:
                typeof event.data === "string"
                  ? event.data.substring(0, MAX_LOG_PREVIEW_LENGTH)
                  : undefined,
            });
            return;
          }

          const parsed = JSON.parse(event.data);

          // Handle pong messages for health checks
          if (parsed && typeof parsed === "object" && parsed.type === "pong") {
            this.handlePong();
            return;
          }

          if (!this.validateMessage(parsed)) {
            logger.warn("Invalid message structure", {
              dataLength,
              preview:
                typeof event.data === "string"
                  ? event.data.substring(0, MAX_LOG_PREVIEW_LENGTH)
                  : undefined,
            });
            return;
          }

          const message = parsed as WebSocketMessage;
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
            dataLength:
              typeof event.data === "string" ? event.data.length : event.data?.byteLength || 0,
            preview:
              typeof event.data === "string"
                ? event.data.substring(0, MAX_LOG_PREVIEW_LENGTH)
                : undefined,
            url: wsUrl,
          });
        }
      };

      this.ws.onerror = (event) => {
        logger.error("Connection error", new Error("WebSocket connection error"), {
          eventType: event.type,
          readyState: this.ws?.readyState,
          url: wsUrl,
          reconnectAttempts: this.reconnectAttempts,
        });
        this.setStatus("error");
      };

      this.ws.onclose = (event) => {
        logger.info("Connection closed", {
          code: event.code,
          reason: event.reason || "No reason provided",
          wasClean: event.wasClean,
          reconnectAttempts: this.reconnectAttempts,
          willReconnect: this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts,
        });
        this.stopPing();
        this.ws = null;
        this.setStatus("disconnected");

        if (
          this.shouldReconnect &&
          !this.isPaused &&
          this.reconnectAttempts < this.maxReconnectAttempts
        ) {
          this.scheduleReconnect();
        } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          logger.warn("Max reconnection attempts reached", {
            maxAttempts: this.maxReconnectAttempts,
          });
          this.setStatus("error");
        }
      };
    } catch (error) {
      logger.error("Failed to create connection", error, {
        url: wsUrl,
        reconnectAttempts: this.reconnectAttempts,
      });
      this.scheduleReconnect();
    }
  }

  /**
   * Validates that a parsed message matches the expected structure.
   * Uses discriminated union type checking.
   */
  private validateMessage(data: unknown): data is WebSocketMessage {
    if (typeof data !== "object" || data === null) {
      return false;
    }
    const msg = data as Record<string, unknown>;
    if (typeof msg.type !== "string") {
      return false;
    }
    const validTypes = ["message", "stats", "pong"];
    if (!validTypes.includes(msg.type)) {
      return false;
    }

    // Type-specific validation
    switch (msg.type) {
      case "message":
        return typeof msg.data === "object" && msg.data !== null;
      case "stats":
        return (
          typeof msg.data === "object" &&
          msg.data !== null &&
          typeof (msg.data as Record<string, unknown>).total === "number" &&
          typeof (msg.data as Record<string, unknown>).byPriority === "object" &&
          typeof (msg.data as Record<string, unknown>).byType === "object"
        );
      case "pong":
        return true; // pong has no data requirement
      default:
        return false;
    }
  }

  private startPing(): void {
    this.stopPing();
    this.pingTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN && !this.isPaused) {
        try {
          this.ws.send(JSON.stringify({ type: "ping" }));
          this.pongTimer = setTimeout(() => {
            logger.warn("Pong timeout, reconnecting");
            if (this.ws) {
              this.ws.close();
            }
          }, PONG_TIMEOUT);
        } catch (error) {
          logger.error("Failed to send ping", error);
        }
      }
    }, PING_INTERVAL);
  }

  private handlePong(): void {
    if (this.pongTimer) {
      clearTimeout(this.pongTimer);
      this.pongTimer = null;
    }
  }

  private stopPing(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
    if (this.pongTimer) {
      clearTimeout(this.pongTimer);
      this.pongTimer = null;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer !== null) {
      return; // Already scheduled
    }

    if (!this.shouldReconnect || this.isPaused) {
      logger.debug("Reconnection skipped because reconnect flag is disabled or paused");
      return;
    }

    this.reconnectAttempts++;
    const baseDelay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.maxReconnectDelay
    );

    // Add jitter (0-30% of delay) to prevent thundering herd
    const jitter = Math.random() * 0.3 * baseDelay;
    const delay = baseDelay + jitter;

    logger.info("Scheduling reconnection", {
      delayMs: delay,
      baseDelay,
      jitter,
      attempt: this.reconnectAttempts,
      maxAttempts: this.maxReconnectAttempts,
    });

    this.setStatus("reconnecting");
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
    this.cleanup();
    this.setStatus("disconnected");
  }

  /**
   * Pauses the WebSocket connection (e.g., when tab is hidden).
   * Stops ping/pong and closes connection, but allows resumption.
   */
  pause(): void {
    if (this.isPaused) return;
    this.isPaused = true;
    logger.info("Pausing WebSocket connection");
    this.stopPing();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Resumes the WebSocket connection (e.g., when tab becomes visible).
   * Reconnects if shouldReconnect is true.
   */
  resume(): void {
    if (!this.isPaused) return;
    this.isPaused = false;
    logger.info("Resuming WebSocket connection");
    if (this.shouldReconnect && !this.isConnected()) {
      this.connect();
    }
  }

  /**
   * Sends a message to the WebSocket server.
   * @param data - Data to send (will be JSON stringified)
   */
  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(data));
      } catch (error) {
        logger.error("Failed to send message", error);
      }
    } else {
      logger.warn("Cannot send message, WebSocket not connected", {
        readyState: this.ws?.readyState,
        status: this.currentStatus,
      });
    }
  }

  private cleanup(): void {
    this.stopPing();
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
    logger.debug("Handler subscribed", { handlerCount: this.handlers.size });
    return () => {
      this.handlers.delete(handler);
      logger.debug("Handler unsubscribed", { handlerCount: this.handlers.size });
    };
  }

  /**
   * Checks if the WebSocket is currently connected.
   * @returns true if connected, false otherwise
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Gets the current connection status.
   * @returns Current WebSocket status
   */
  getStatus(): WebSocketStatus {
    return this.currentStatus;
  }

  /**
   * Subscribes to status changes.
   * @param handler - Function to call when status changes
   * @returns Unsubscribe function
   */
  onStatusChange(handler: WebSocketStatusHandler): () => void {
    this.statusHandlers.add(handler);
    // Immediately notify of current status
    try {
      handler(this.currentStatus);
    } catch (error) {
      logger.error("Status handler error on subscribe", error);
    }
    return () => {
      this.statusHandlers.delete(handler);
    };
  }

  /**
   * Gets the current number of reconnection attempts.
   * @returns Number of reconnection attempts
   */
  getReconnectAttempts(): number {
    return this.reconnectAttempts;
  }

  /**
   * Gets the maximum number of reconnection attempts.
   * @returns Maximum number of reconnection attempts
   */
  getMaxReconnectAttempts(): number {
    return this.maxReconnectAttempts;
  }

  /**
   * Destroys the WebSocket client and cleans up all resources.
   * Should be called when the client is no longer needed.
   */
  destroy(): void {
    this.disconnect();
    if (this.visibilityHandler && typeof document !== "undefined") {
      document.removeEventListener("visibilitychange", this.visibilityHandler);
      this.visibilityHandler = null;
    }
    this.handlers.clear();
    this.statusHandlers.clear();
  }

  // Test helper methods (only used in tests)
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
