// Svelte store for WebSocket connection status

import { writable } from "svelte/store";
import { wsClient, type WebSocketMessage } from "../services/websocket.ts";
import { messages } from "./messages.ts";
import { stats } from "./stats.ts";

function createWebSocketStore() {
  const { subscribe, set } = writable<boolean>(false);

  // Update store when connection status changes
  const checkConnection = () => {
    set(wsClient.isConnected());
  };

  // Check connection status periodically
  let interval: number | null = null;

  // Subscribe to WebSocket messages and update stores
  // Store the unsubscribe function to prevent memory leaks
  let unsubscribe: (() => void) | null = wsClient.subscribe((message: WebSocketMessage) => {
    if (message.type === "message") {
      messages.add(message.data);
    } else if (message.type === "stats-total") {
      stats.setTotal(message.data.total);
    } else if (message.type === "stats-priority") {
      stats.setByPriority(message.data.byPriority);
    } else if (message.type === "stats-type") {
      stats.setByType(message.data.byType);
    }
  });

  return {
    subscribe,
    connect: () => {
      wsClient.connect();
      checkConnection();
      interval = setInterval(checkConnection, 1000);
    },
    disconnect: () => {
      wsClient.disconnect();
      set(false);
      if (interval !== null) {
        clearInterval(interval);
        interval = null;
      }
    },
    isConnected: () => wsClient.isConnected(),
    cleanup: () => {
      if (interval !== null) {
        clearInterval(interval);
        interval = null;
      }
      wsClient.disconnect();
      // Unsubscribe from WebSocket messages to prevent memory leaks
      if (unsubscribe !== null) {
        unsubscribe();
        unsubscribe = null;
      }
    },
  };
}

export const websocket = createWebSocketStore();

