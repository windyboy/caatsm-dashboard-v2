import type {
  WSMessage,
  WSStatsData,
  WSStatsDeltaData,
  WSHealthData,
  Telegram,
  WSConnectionStatus,
} from "$lib/types";
import { WSMessageType } from "$lib/types";

type StatsHandler = (data: WSStatsData) => void;
type StatsDeltaHandler = (data: WSStatsDeltaData) => void;
type HealthHandler = (data: WSHealthData) => void;
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
  private statsDeltaHandlers: StatsDeltaHandler[] = [];
  private healthHandlers: HealthHandler[] = [];
  private telegramHandlers: TelegramHandler[] = [];
  private statusListeners: ((status: WSConnectionStatus) => void)[] = [];

  private _status: WSConnectionStatus = "disconnected";
  private _url: string;

  constructor(url?: string) {
    // Don't call getWebSocketUrl() here - defer until connect() when window is available
    this._url = url || "";
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
    // Check if we're in browser environment
    if (typeof window === "undefined") {
      throw new Error("WebSocket URL can only be determined in browser environment");
    }

    if (import.meta.env.DEV) {
      // Use relative path in dev mode, Vite proxy will forward to backend
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      return `${protocol}//${window.location.host}/ws`;
    }
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    return `${protocol}//${host}/ws`;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    // Get URL when connecting (browser environment guaranteed)
    if (!this._url) {
      this._url = this.getWebSocketUrl();
      if (!this._url) {
        this.setStatus("error");
        return;
      }
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
          // Silently handle parse errors
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
      case WSMessageType.STATS_FULL:
        try {
          const statsData = message.data as WSStatsData;
          this.statsHandlers.forEach((handler) => handler(statsData));
        } catch (err) {
          // Silently handle handler errors
        }
        break;

      case WSMessageType.STATS_DELTA:
        try {
          const deltaData = message.data as WSStatsDeltaData;
          this.statsDeltaHandlers.forEach((handler) => handler(deltaData));
        } catch (err) {
          // Silently handle handler errors
        }
        break;

      case WSMessageType.HEALTH:
        try {
          const healthData = message.data as WSHealthData;
          this.healthHandlers.forEach((handler) => handler(healthData));
        } catch (err) {
          // Silently handle handler errors
        }
        break;

      case WSMessageType.MESSAGE:
        try {
          const telegram = message.data as Telegram;
          this.telegramHandlers.forEach((handler) => handler(telegram));
        } catch (err) {
          // Silently handle handler errors
        }
        break;

      default:
        // Unknown message type, ignore
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

  onStatsDelta(handler: StatsDeltaHandler): () => void {
    this.statsDeltaHandlers.push(handler);
    return () => {
      const index = this.statsDeltaHandlers.indexOf(handler);
      if (index > -1) this.statsDeltaHandlers.splice(index, 1);
    };
  }

  onHealth(handler: HealthHandler): () => void {
    this.healthHandlers.push(handler);
    return () => {
      const index = this.healthHandlers.indexOf(handler);
      if (index > -1) this.healthHandlers.splice(index, 1);
    };
  }

  onMessage(handler: TelegramHandler): () => void {
    this.telegramHandlers.push(handler);
    return () => {
      const index = this.telegramHandlers.indexOf(handler);
      if (index > -1) {
        this.telegramHandlers.splice(index, 1);
      }
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

