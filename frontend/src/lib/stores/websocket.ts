// Svelte store for WebSocket connection status
// Note: Subscription is set up in connect() to avoid SSR issues.
// In SvelteKit, module-level stores are safe for SSR as each request gets a fresh module context.

import { derived, writable } from "svelte/store";
import { type WebSocketStatus, wsClient } from "../services/websocket";
import type { WebSocketMessage } from "../utils/types";
import { messages } from "./data/messages";
import { stats } from "./data/stats";
import { createLogger } from "../utils/logger";
import { search } from "../services/api";

const logger = createLogger("WebSocketStore");

function createWebSocketStore() {
  const status = writable<WebSocketStatus>("disconnected");
  const error = writable<string | null>(null);
  const reconnectAttempts = writable<number>(0);

  // Derived store for boolean connection state (backward compatibility)
  const isConnected = derived(status, ($status) => $status === "connected");

  // Subscribe to WebSocket messages and update stores
  // Store the unsubscribe functions to prevent memory leaks
  let messageUnsubscribe: (() => void) | null = null;
  let statusUnsubscribe: (() => void) | null = null;
  
  // Track last message timestamp for resync after reconnect
  let lastMessageTime: string | null = null;
  let previousStatus: WebSocketStatus = "disconnected";
  
  /**
   * Resyncs messages after reconnection by fetching missed messages.
   * Uses the last message timestamp to fetch messages since disconnection.
   */
  async function resyncAfterReconnect(): Promise<void> {
    if (!lastMessageTime) {
      logger.debug("No last message time, skipping resync");
      return; // No messages yet, nothing to resync
    }

    try {
      logger.info("Resyncing messages after reconnect", { after: lastMessageTime });
      const result = await search({
        start_time: lastMessageTime,
        limit: 100, // Limit to last 100 messages to avoid overwhelming UI
        sort_by: "time",
        order: "desc",
      });

      if (result.telegrams && result.telegrams.length > 0) {
        // Messages are fetched with order: "desc" (newest first)
        // Since UI displays newest first (at top), we keep them in desc order
        // addMultiple will handle deduplication and add them at the beginning
        messages.addMultiple(result.telegrams);
        logger.info("Resync complete", { count: result.telegrams.length });
      } else {
        logger.debug("No messages found during resync");
      }
    } catch (error) {
      logger.error("Resync failed", error);
      // Don't throw - resync failure shouldn't break the connection
    }
  }

  return {
    status: { subscribe: status.subscribe },
    error: { subscribe: error.subscribe },
    reconnectAttempts: { subscribe: reconnectAttempts.subscribe },
    isConnected: { subscribe: isConnected.subscribe },

    connect: () => {
      logger.info("Connecting WebSocket store", {
        currentStatus: wsClient.getStatus(),
      });

      // Only subscribe if we're in a browser environment and not already subscribed
      if (typeof window !== "undefined") {
        // Subscribe to status changes (event-driven, no polling)
        if (statusUnsubscribe === null) {
          statusUnsubscribe = wsClient.onStatusChange((newStatus) => {
            logger.debug("WebSocket status changed", { status: newStatus, previousStatus });
            status.set(newStatus);

            // Update reconnect attempts
            reconnectAttempts.set(wsClient.getReconnectAttempts());

            // Trigger resync when reconnecting after a disconnect
            if (newStatus === "connected" && previousStatus !== "connected") {
              // Just reconnected - trigger resync
              resyncAfterReconnect();
            }

            // Clear error on successful connection
            if (newStatus === "connected") {
              error.set(null);
            } else if (newStatus === "error") {
              const attempts = wsClient.getReconnectAttempts();
              const maxAttempts = wsClient.getMaxReconnectAttempts();
              if (attempts >= maxAttempts) {
                error.set("Unable to connect to live stream after multiple attempts. Please refresh the page or check your network connection.");
              } else {
                error.set("Connection lost. Attempting to reconnect...");
              }
            } else if (newStatus === "reconnecting") {
              const attempts = wsClient.getReconnectAttempts();
              const maxAttempts = wsClient.getMaxReconnectAttempts();
              error.set(`Reconnecting... (attempt ${attempts + 1}/${maxAttempts})`);
            } else if (newStatus === "connecting") {
              error.set("Connecting to live stream...");
            } else if (newStatus === "disconnected") {
              error.set("Disconnected from live stream.");
            }
            
            // Update previous status for next comparison
            previousStatus = newStatus;
          });
        }

        // Subscribe to WebSocket messages
        if (messageUnsubscribe === null) {
          messageUnsubscribe = wsClient.subscribe((message: WebSocketMessage) => {
            logger.debug("Received WebSocket message", {
              type: message.type,
            });

            // Use discriminated union for type-safe handling
            switch (message.type) {
              case "message":
                // Track timestamp for resync
                if (message.data.time) {
                  lastMessageTime = message.data.time;
                }
                messages.add(message.data);
                break;
              case "stats":
                // Update statistics from WebSocket message
                stats.setStats({
                  total: message.data.total,
                  byPriority: message.data.byPriority,
                  byType: message.data.byType,
                });
                break;
              case "pong":
                // Ping/pong handled internally by client
                break;
            }
          });
        }

        wsClient.connect();
      }
    },

    disconnect: () => {
      logger.info("Disconnecting WebSocket store");
      wsClient.disconnect();
      status.set("disconnected");
      error.set(null);
      reconnectAttempts.set(0);

      // Unsubscribe when disconnecting
      if (messageUnsubscribe !== null) {
        messageUnsubscribe();
        messageUnsubscribe = null;
      }
      if (statusUnsubscribe !== null) {
        statusUnsubscribe();
        statusUnsubscribe = null;
      }
    },

    send: (data: unknown) => {
      wsClient.send(data);
    },

    checkIsConnected: () => wsClient.isConnected(),

    cleanup: () => {
      logger.info("Cleaning up WebSocket store");
      wsClient.disconnect();
      status.set("disconnected");
      error.set(null);
      reconnectAttempts.set(0);

      // Unsubscribe from WebSocket messages to prevent memory leaks
      if (messageUnsubscribe !== null) {
        messageUnsubscribe();
        messageUnsubscribe = null;
      }
      if (statusUnsubscribe !== null) {
        statusUnsubscribe();
        statusUnsubscribe = null;
      }
    },
  };
}

export const websocket = createWebSocketStore();
