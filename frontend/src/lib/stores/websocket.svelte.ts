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
  private _isInitialBatch = $state(false);
  private _initialBatchCount = $state(0);

  constructor() {
    this.service = new WebSocketService();
    this.service.onStatusChange((status) => {
      const wasConnected = this._status === "connected";
      this._status = status;

      // Handle reconnection: mark as initial batch when reconnecting
      if (status === "connected" && !wasConnected) {
        this._isInitialBatch = true;
        this._initialBatchCount = 0;
      }
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
      const isDuplicate = this.isDuplicateMessage(telegram);

      // Skip if duplicate
      if (isDuplicate) {
        return;
      }

      // Track initial batch messages (first 50 messages after connection)
      if (this._isInitialBatch) {
        this._initialBatchCount++;
        // After receiving initial batch (approximately 50 messages), stop tracking
        if (this._initialBatchCount >= 50) {
          this._isInitialBatch = false;
        }
      }

      // Add new message at the beginning (newest first)
      // Limit to 100 messages to prevent memory leaks
      const updated = [telegram, ...this._newMessages];
      this._newMessages = updated.slice(0, 100);
    });
  }

  // isDuplicateMessage checks if a telegram is a duplicate of existing messages
  private isDuplicateMessage(telegram: Telegram): boolean {
    return this._newMessages.some((msg) => {
      // Primary check: If both have message_id, compare by ID
      if (msg.message_id && telegram.message_id) {
        return msg.message_id === telegram.message_id;
      }

      // Secondary check: Compare by content+time+type
      // Handle empty strings properly - both must be empty or both must match
      const contentMatch =
        (msg.content === undefined || msg.content === null || msg.content === "") &&
        (telegram.content === undefined || telegram.content === null || telegram.content === "")
          ? true
          : msg.content === telegram.content;

      const timeMatch =
        (msg.time === undefined || msg.time === null || msg.time === "") &&
        (telegram.time === undefined || telegram.time === null || telegram.time === "")
          ? true
          : msg.time === telegram.time;

      const typeMatch =
        (msg.type === undefined || msg.type === null || msg.type === "") &&
        (telegram.type === undefined || telegram.type === null || telegram.type === "")
          ? true
          : msg.type === telegram.type;

      return contentMatch && timeMatch && typeMatch;
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
      postgresql: health.postgresql ? (health.postgresql as HealthSnapshot["postgresql"]) : undefined,
      meilisearch: health.meilisearch
        ? (health.meilisearch as HealthSnapshot["meilisearch"])
        : undefined,
      redis: health.redis ? (health.redis as HealthSnapshot["redis"]) : undefined,
      nats: health.nats ? (health.nats as HealthSnapshot["nats"]) : undefined,
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
    this._isInitialBatch = false;
    this._initialBatchCount = 0;
  }

  // Mark that initial batch is complete (useful for external callers)
  markInitialBatchComplete(): void {
    this._isInitialBatch = false;
    this._initialBatchCount = 0;
  }
}

let storeInstance: WebSocketStore | null = null;

export function getWebSocketStore(): WebSocketStore {
  if (!storeInstance) {
    storeInstance = new WebSocketStore();
  }
  return storeInstance;
}
