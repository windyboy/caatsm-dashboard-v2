import type { WSMessage, WSStatsData, Telegram, WSConnectionStatus } from "$lib/types";
import { WSMessageType } from "$lib/types";

type StatsHandler = (data: WSStatsData) => void;
type TelegramHandler = (data: Telegram) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private maxReconnectDelay = 30000;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private shouldReconnect = true;
  private statsHandlers: StatsHandler[] = [];
  private telegramHandlers: TelegramHandler[] = [];
  private statusListeners: ((status: WSConnectionStatus) => void)[] = [];

  private _status: WSConnectionStatus = "disconnected";
  private _url: string;

  constructor(url?: string) {
    this._url = url || this.getWebSocketUrl();
  }

  get status(): WSConnectionStatus {
    return this._status;
  }

  onStatusChange(listener: (status: WSConnectionStatus) => void): () => void {
    this.statusListeners.push(listener);
    return () => {
      const index = this.statusListeners.indexOf(listener);
      if (index > -1) this.statusListeners.splice(index, 1);
    };
  }

  private setStatus(status: WSConnectionStatus): void {
    if (this._status !== status) {
      this._status = status;
      this.statusListeners.forEach((listener) => listener(status));
    }
  }

  private getWebSocketUrl(): string {
    if (import.meta.env.DEV) {
      return "ws://localhost:5173/ws";
    }
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    return `${protocol}//${host}/ws`;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.setStatus("connecting");
    this.shouldReconnect = true;

    try {
      this.ws = new WebSocket(this._url);

      this.ws.onopen = () => {
        this.setStatus("connected");
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (err) {
          if (import.meta.env.DEV) {
            console.error("Failed to parse WebSocket message:", err);
          }
        }
      };

      this.ws.onerror = () => {
        this.setStatus("error");
      };

      this.ws.onclose = () => {
        this.setStatus("disconnected");
        this.ws = null;

        if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.scheduleReconnect();
        }
      };
    } catch (err) {
      this.setStatus("error");
      if (import.meta.env.DEV) {
        console.error("WebSocket connection error:", err);
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.maxReconnectDelay
    );

    this.reconnectTimer = setTimeout(() => {
      if (this.shouldReconnect) {
        this.connect();
      }
    }, delay);
  }

  private handleMessage(message: WSMessage): void {
    switch (message.type) {
      case WSMessageType.STATS:
        try {
          const statsData = message.data as WSStatsData;
          this.statsHandlers.forEach((handler) => handler(statsData));
        } catch (err) {
          if (import.meta.env.DEV) {
            console.error("Failed to handle stats message:", err);
          }
        }
        break;

      case WSMessageType.MESSAGE:
        try {
          const telegram = message.data as Telegram;
          this.telegramHandlers.forEach((handler) => handler(telegram));
        } catch (err) {
          if (import.meta.env.DEV) {
            console.error("Failed to handle message:", err);
          }
        }
        break;
    }
  }

  onStats(handler: StatsHandler): () => void {
    this.statsHandlers.push(handler);
    return () => {
      const index = this.statsHandlers.indexOf(handler);
      if (index > -1) this.statsHandlers.splice(index, 1);
    };
  }

  onMessage(handler: TelegramHandler): () => void {
    this.telegramHandlers.push(handler);
    return () => {
      const index = this.telegramHandlers.indexOf(handler);
      if (index > -1) this.telegramHandlers.splice(index, 1);
    };
  }

  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.setStatus("disconnected");
  }
}

