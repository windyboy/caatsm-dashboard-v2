// Svelte store for WebSocket connection status
// Note: Subscription is set up in connect() to avoid SSR issues.
// In SvelteKit, module-level stores are safe for SSR as each request gets a fresh module context.

import { writable, derived } from "svelte/store";
import { wsClient, type WebSocketStatus } from "../services/websocket.ts";
import type { WebSocketMessage } from "../utils/types.ts";
import { messages } from "./messages.ts";
import { stats } from "./stats.ts";
import { createLogger } from "../utils/logger.ts";

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
            logger.debug("WebSocket status changed", { status: newStatus });
            status.set(newStatus);
            
            // Update reconnect attempts
            reconnectAttempts.set(wsClient.getReconnectAttempts());
            
            // Clear error on successful connection
            if (newStatus === "connected") {
              error.set(null);
            } else if (newStatus === "error") {
              error.set("Connection error. Please check your network.");
            }
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
                messages.add(message.data);
                break;
              case "stats-total":
                stats.setTotal(message.data.total);
                break;
              case "stats-priority":
                stats.setByPriority(message.data.byPriority);
                break;
              case "stats-type":
                stats.setByType(message.data.byType);
                break;
              case "pong":
                // Ping/pong handled internally by client
                break;
            }
          });
        }
      }
      
      wsClient.connect();
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
