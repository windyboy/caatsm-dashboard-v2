import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { WebSocketClient } from "../../src/lib/services/websocket";

describe("WebSocketClient", () => {
  let client: WebSocketClient;

  beforeEach(() => {
    client = new WebSocketClient();
  });

  afterEach(() => {
    client.disconnect();
    vi.clearAllTimers();
  });

  it("should connect successfully", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should not connect if already connected", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);
      const initialHandlerCount = client.getHandlerCount();
      client.connect();
      await vi.advanceTimersByTimeAsync(0);
      expect(client.isConnected()).toBe(true);
      expect(client.getHandlerCount()).toBe(initialHandlerCount);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should parse and handle messages", async () => {
    vi.useFakeTimers();
    try {
      const handler = vi.fn();
      client.subscribe(handler);

      client.connect();
      await vi.advanceTimersByTimeAsync(10);

      const message = {
        type: "message",
        data: { message_id: "TEST-001" },
      };

      const ws = (client as any).ws;
      if (ws && ws.onmessage) {
        ws.onmessage({
          data: JSON.stringify(message),
        } as MessageEvent);
      }

      expect(handler).toHaveBeenCalledWith(message);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle invalid JSON messages", async () => {
    vi.useFakeTimers();
    try {
      const handler = vi.fn();
      client.subscribe(handler);

      client.connect();
      await vi.advanceTimersByTimeAsync(10);

      const ws = (client as any).ws;
      if (ws && ws.onmessage) {
        ws.onmessage({
          data: "invalid json",
        } as MessageEvent);
      }

      expect(handler).not.toHaveBeenCalled();
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle reconnection with exponential backoff", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(20);
      expect(client.isConnected()).toBe(true);

      const ws = (client as any).ws;
      if (ws && ws.onclose) {
        ws.onclose({
          code: 1000,
          reason: "test",
          wasClean: false,
        } as CloseEvent);
      }

      // Wait for scheduleReconnect to complete
      await vi.advanceTimersByTimeAsync(0);
      expect(client.getReconnectAttempts()).toBe(1);

      await vi.advanceTimersByTimeAsync(1000);

      expect(client.getReconnectAttempts()).toBeGreaterThan(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should stop reconnecting after max attempts", async () => {
    vi.useFakeTimers();
    const OriginalWebSocket = (globalThis as any).WebSocket;
    (globalThis as any).WebSocket = class FailingWebSocket {
      static CONNECTING = 0;
      static OPEN = 1;
      static CLOSING = 2;
      static CLOSED = 3;
      readyState = FailingWebSocket.CLOSED;
      url: string;
      onopen: ((event: Event) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onclose: ((event: CloseEvent) => void) | null = null;

      constructor(url: string) {
        this.url = url;
        setTimeout(() => {
          if (this.onclose) {
            this.onclose(new CloseEvent("close", { code: 1006, wasClean: false }));
          }
        }, 0);
      }

      send() {}
      close() {}
    } as any;

    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(0);

      const maxAttempts = 10;

      for (let i = 0; i < maxAttempts; i++) {
        const delay = Math.min(1000 * Math.pow(2, i), 30000);
        await vi.advanceTimersByTimeAsync(delay);
        await vi.advanceTimersByTimeAsync(0);
      }

      await vi.runAllTimersAsync();

      const finalAttempts = client.getReconnectAttempts();
      expect(finalAttempts).toBeGreaterThanOrEqual(maxAttempts);
    } finally {
      (globalThis as any).WebSocket = OriginalWebSocket;
      vi.useRealTimers();
    }
  });

  it("should subscribe and unsubscribe handlers", () => {
    const handler1 = vi.fn();
    const handler2 = vi.fn();

    const unsubscribe1 = client.subscribe(handler1);
    const unsubscribe2 = client.subscribe(handler2);

    expect(client.getHandlerCount()).toBe(2);

    unsubscribe1();
    expect(client.getHandlerCount()).toBe(1);

    unsubscribe2();
    expect(client.getHandlerCount()).toBe(0);
  });

  it("should disconnect properly", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(20);
      expect(client.isConnected()).toBe(true);

      client.disconnect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(false);
      expect(client.getShouldReconnect()).toBe(false);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should reset reconnection state on successful connection", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(20);
      expect(client.getReconnectAttempts()).toBe(0);
      expect(client.getReconnectDelay()).toBe(1000);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle concurrent connect calls", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      client.connect();
      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle handler that throws exception", async () => {
    vi.useFakeTimers();
    try {
      const handler1 = vi.fn(() => {
        throw new Error("Handler error");
      });
      const handler2 = vi.fn();
      client.subscribe(handler1);
      client.subscribe(handler2);

      client.connect();
      await vi.advanceTimersByTimeAsync(10);

      const message = {
        type: "message",
        data: { message_id: "TEST-001" },
      };

      const ws = (client as any).ws;
      if (ws && ws.onmessage) {
        ws.onmessage({
          data: JSON.stringify(message),
        } as MessageEvent);
      }

      expect(handler1).toHaveBeenCalled();
      expect(handler2).toHaveBeenCalled();
    } finally {
      vi.useRealTimers();
    }
  });

  it("should respect max reconnect delay cap", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);

      const ws = (client as any).ws;
      for (let i = 0; i < 15; i++) {
        if (ws && ws.onclose) {
          ws.onclose({
            code: 1000,
            reason: "test",
            wasClean: false,
          } as CloseEvent);
        }
        await vi.advanceTimersByTimeAsync(0);
        const delay = Math.min(1000 * Math.pow(2, i), 30000);
        await vi.advanceTimersByTimeAsync(delay);
      }

      expect(client.getReconnectDelay()).toBeLessThanOrEqual(30000);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should not reconnect after explicit disconnect", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);

      client.disconnect();
      await vi.advanceTimersByTimeAsync(0);

      const ws = (client as any).ws;
      if (ws && ws.onclose) {
        ws.onclose({
          code: 1000,
          reason: "test",
          wasClean: false,
        } as CloseEvent);
      }

      await vi.advanceTimersByTimeAsync(5000);
      expect(client.getShouldReconnect()).toBe(false);
      expect(client.isConnected()).toBe(false);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle connection during reconnect timer", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);

      const ws = (client as any).ws;
      if (ws && ws.onclose) {
        ws.onclose({
          code: 1000,
          reason: "test",
          wasClean: false,
        } as CloseEvent);
      }

      await vi.advanceTimersByTimeAsync(0);
      expect(client.getReconnectAttempts()).toBeGreaterThan(0);

      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle empty message data", async () => {
    vi.useFakeTimers();
    try {
      const handler = vi.fn();
      client.subscribe(handler);

      client.connect();
      await vi.advanceTimersByTimeAsync(10);

      const ws = (client as any).ws;
      if (ws && ws.onmessage) {
        // null and undefined data should be rejected by validation
        ws.onmessage({
          data: JSON.stringify({ type: "message", data: null }),
        } as MessageEvent);
        // Note: undefined gets removed during JSON.stringify, so this becomes { type: "message" }
        ws.onmessage({
          data: JSON.stringify({ type: "message" }),
        } as MessageEvent);
      }

      // Both messages should be rejected: null fails validation, missing data fails validation
      expect(handler).toHaveBeenCalledTimes(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle very large messages", async () => {
    vi.useFakeTimers();
    try {
      const handler = vi.fn();
      client.subscribe(handler);

      client.connect();
      await vi.advanceTimersByTimeAsync(10);

      const largeData = "x".repeat(1048577);
      const ws = (client as any).ws;
      if (ws && ws.onmessage) {
        ws.onmessage({
          data: JSON.stringify({ type: "message", data: largeData }),
        } as MessageEvent);
      }

      expect(handler).not.toHaveBeenCalled();
    } finally {
      vi.useRealTimers();
    }
  });

  it("should handle multiple disconnect calls", async () => {
    vi.useFakeTimers();
    try {
      client.connect();
      await vi.advanceTimersByTimeAsync(10);
      expect(client.isConnected()).toBe(true);

      client.disconnect();
      client.disconnect();
      client.disconnect();

      await vi.advanceTimersByTimeAsync(0);
      expect(client.isConnected()).toBe(false);
      expect(client.getShouldReconnect()).toBe(false);
    } finally {
      vi.useRealTimers();
    }
  });

  it("should validate message type enum", async () => {
    vi.useFakeTimers();
    try {
      const handler = vi.fn();
      client.subscribe(handler);

      client.connect();
      await vi.advanceTimersByTimeAsync(10);

      const ws = (client as any).ws;
      if (ws && ws.onmessage) {
        // Invalid type should be rejected
        ws.onmessage({
          data: JSON.stringify({ type: "invalid-type", data: {} }),
        } as MessageEvent);
        // Valid message type with valid data structure
        ws.onmessage({
          data: JSON.stringify({ type: "message", data: {} }),
        } as MessageEvent);
        // stats requires data.total, byPriority, and byType to be present and correct types
        ws.onmessage({
          data: JSON.stringify({ type: "stats", data: {} }),
        } as MessageEvent);
      }

      // Only the "message" type with valid data should pass validation
      expect(handler).toHaveBeenCalledTimes(1);
      expect(handler).toHaveBeenCalledWith(expect.objectContaining({ type: "message" }));
      expect(handler).not.toHaveBeenCalledWith(expect.objectContaining({ type: "invalid-type" }));
      expect(handler).not.toHaveBeenCalledWith(expect.objectContaining({ type: "stats" }));
    } finally {
      vi.useRealTimers();
    }
  });
});
