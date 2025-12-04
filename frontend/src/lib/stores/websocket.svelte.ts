import { WebSocketService } from "$lib/services/websocket";
import type {
  WSConnectionStatus,
  Telegram,
  WSStatsData,
  WSStatsDeltaData,
  WSHealthData,
  HealthSnapshot,
} from "$lib/types";

class WebSocketStore {
  private service: WebSocketService;
  private _status = $state<WSConnectionStatus>("disconnected");
  private _stats = $state<WSStatsData | null>(null);
  private _health = $state<HealthSnapshot | null>(null);
  private _newMessages = $state<Telegram[]>([]);

  constructor() {
    this.service = new WebSocketService();
    this.service.onStatusChange((status) => {
      this._status = status;
    });

    // Handle full stats (replaces current stats)
    this.service.onStats((stats) => {
      this._stats = stats;
    });

    // Handle incremental stats (accumulates)
    this.service.onStatsDelta((delta) => {
      this.applyStatsDelta(delta);
    });

    // Handle health updates
    this.service.onHealth((health) => {
      this._health = this.convertHealthData(health);
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

  // applyStatsDelta applies incremental stats updates to current stats
  private applyStatsDelta(delta: WSStatsDeltaData): void {
    if (!this._stats) {
      // If no stats exist, initialize from delta
      this._stats = {
        total: delta.total,
        byType: { ...delta.byType },
        activeRoutes: 0,
        messagesPerSec: 0,
        timeWindow: delta.timeWindow,
      };
      return;
    }

    // Accumulate totals
    this._stats.total = (this._stats.total || 0) + delta.total;

    // Accumulate by type
    if (delta.byType) {
      if (!this._stats.byType) {
        this._stats.byType = {};
      }
      for (const [type, count] of Object.entries(delta.byType)) {
        this._stats.byType[type] = (this._stats.byType[type] || 0) + count;
      }
    }

    // Update time window if provided
    if (delta.timeWindow) {
      this._stats.timeWindow = delta.timeWindow;
    }
  }

  // convertHealthData converts WSHealthData to HealthSnapshot format
  private convertHealthData(health: WSHealthData): HealthSnapshot {
    return {
      status: health.status,
      version: health.version,
      timestamp: health.timestamp,
      postgresql: health.postgresql as any,
      meilisearch: health.meilisearch as any,
      redis: health.redis as any,
      nats: health.nats as any,
    };
  }

  get status(): WSConnectionStatus {
    return this._status;
  }

  get stats(): WSStatsData | null {
    return this._stats;
  }

  get health(): HealthSnapshot | null {
    return this._health;
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
