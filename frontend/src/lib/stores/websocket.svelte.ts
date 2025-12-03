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
      this._newMessages = [telegram, ...this._newMessages];
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

