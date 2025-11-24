// WebSocket client for real-time updates

import { createLogger } from "../utils/logger.ts";

const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:3002/ws";
const logger = createLogger("WebSocket");

export interface WebSocketMessage {
  type: "message" | "stats-total" | "stats-priority" | "stats-type";
  data: any;
}

export type WebSocketMessageHandler = (message: WebSocketMessage) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000; // Start with 1 second
  private maxReconnectDelay = 30000; // Max 30 seconds
  private handlers: Set<WebSocketMessageHandler> = new Set();
  private shouldReconnect = true;
  private reconnectTimer: number | null = null;

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
        this.reconnectDelay = 1000;
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          logger.debug("Message received", {
            type: message.type,
            dataSize: JSON.stringify(message.data).length,
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
            dataLength: event.data?.length,
            dataPreview: typeof event.data === "string" ? event.data.substring(0, 100) : undefined,
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

  subscribe(handler: WebSocketMessageHandler): () => void {
    this.handlers.add(handler);
    return () => {
      this.handlers.delete(handler);
    };
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
export const wsClient = new WebSocketClient();

