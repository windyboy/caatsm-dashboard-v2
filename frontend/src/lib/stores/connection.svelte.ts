import { getWebSocketStore } from "./websocket.svelte";

type ConnectionStatus = "connected" | "connecting" | "disconnected" | "degraded";

class ConnectionStore {
  private wsStore: ReturnType<typeof getWebSocketStore> | null = null;
  private _httpStatus = $state<ConnectionStatus>("disconnected");
  private _lastHttpError = $state<string | null>(null);
  private _retryCount = $state(0);
  private _nextRetryAt = $state<Date | null>(null);
  private retryTimer: ReturnType<typeof setTimeout> | null = null;
  private maxRetries = 5;
  private baseRetryDelay = 2000; // 2 seconds
  private maxRetryDelay = 30000; // 30 seconds

  constructor() {
    // Initialize WebSocket store reference
    // Note: We can't use $effect in constructor, so we'll access it lazily
    // The status will be accessed through getters which will trigger reactivity
  }

  private getWsStore(): ReturnType<typeof getWebSocketStore> {
    if (!this.wsStore) {
      this.wsStore = getWebSocketStore();
    }
    return this.wsStore;
  }

  get httpStatus(): ConnectionStatus {
    return this._httpStatus;
  }

  get wsStatus() {
    return this.getWsStore().status;
  }

  get overallStatus(): ConnectionStatus {
    const wsConnected = this.wsStatus === "connected";
    const httpConnected = this._httpStatus === "connected";

    if (wsConnected && httpConnected) return "connected";
    if (wsConnected || httpConnected) return "degraded";
    if (this.wsStatus === "connecting" || this._httpStatus === "connecting") return "connecting";
    return "disconnected";
  }

  get lastError(): string | null {
    return this._lastHttpError;
  }

  get retryCount(): number {
    return this._retryCount;
  }

  get nextRetryAt(): Date | null {
    return this._nextRetryAt;
  }

  get isRetrying(): boolean {
    return this.retryTimer !== null;
  }

  markHttpSuccess(): void {
    this._httpStatus = "connected";
    this._lastHttpError = null;
    this._retryCount = 0;
    this._nextRetryAt = null;
    this.cancelRetry();
  }

  markHttpError(error: string): void {
    this._httpStatus = "disconnected";
    this._lastHttpError = error;
    this.scheduleRetry();
  }

  private scheduleRetry(): void {
    if (this._retryCount >= this.maxRetries) {
      this._nextRetryAt = null;
      return;
    }

    this.cancelRetry();
    this._retryCount++;

    const delay = Math.min(
      this.baseRetryDelay * Math.pow(2, this._retryCount - 1),
      this.maxRetryDelay
    );

    this._nextRetryAt = new Date(Date.now() + delay);
    this._httpStatus = "connecting";

    this.retryTimer = setTimeout(() => {
      this.retryTimer = null;
      this._nextRetryAt = null;
    }, delay);
  }

  private cancelRetry(): void {
    if (this.retryTimer) {
      clearTimeout(this.retryTimer);
      this.retryTimer = null;
    }
  }

  manualRetry(): void {
    this._retryCount = 0;
    this.cancelRetry();
    this._httpStatus = "connecting";
    this._nextRetryAt = null;
  }
}

let storeInstance: ConnectionStore | null = null;

export function getConnectionStore(): ConnectionStore {
  if (!storeInstance) {
    storeInstance = new ConnectionStore();
  }
  return storeInstance;
}
