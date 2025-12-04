import { WebSocketService } from "$lib/services/websocket";
import type { WSConnectionStatus, Telegram, WSStatsData } from "$lib/types";

class WebSocketStore {
  private service: WebSocketService;
  private _status = $state<WSConnectionStatus>("disconnected");
  private _stats = $state<WSStatsData | null>(null);
  private _newMessages = $state<Telegram[]>([]);

  constructor() {
    this.service = new WebSocketService();
    this.service.onStatusChange((status) => {
      this._status = status;
    });
    this.service.onStats((stats) => {
      this._stats = stats;
    });
    this.service.onMessage((telegram) => {
      // Check if message already exists before adding
      const isDuplicate = this._newMessages.some((msg) => {
        // If both have message_id, compare by ID
        if (msg.message_id && telegram.message_id) {
          return msg.message_id === telegram.message_id;
        }
        // Otherwise, compare by content+time+type
        return (
          msg.content === telegram.content &&
          msg.time === telegram.time &&
          msg.type === telegram.type
        );
      });

      // Skip if duplicate
      if (isDuplicate) {
        return;
      }

      // Add new message at the beginning (newest first)
      // Limit to 100 messages to prevent memory leaks
      const updated = [telegram, ...this._newMessages];
      this._newMessages = updated.slice(0, 100);
    });
  }

  get status(): WSConnectionStatus {
    return this._status;
  }

  get stats(): WSStatsData | null {
    return this._stats;
  }

  get newMessages(): Telegram[] {
    return this._newMessages;
  }

  connect(): void {
    this.service.connect();
  }

  disconnect(): void {
    this.service.disconnect();
  }

  clearNewMessages(): void {
    this._newMessages = [];
  }
}

let storeInstance: WebSocketStore | null = null;

export function getWebSocketStore(): WebSocketStore {
  if (!storeInstance) {
    storeInstance = new WebSocketStore();
  }
  return storeInstance;
}

