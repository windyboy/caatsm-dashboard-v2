import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { WebSocketClient } from "../../src/lib/services/websocket";

interface WebSocketClientTestAccess {
  ws: WebSocket | null;
}

function getWebSocket(client: WebSocketClient): WebSocket | null {
  return (client as unknown as WebSocketClientTestAccess).ws;
}

describe("WebSocketClient", () => {
  let client: WebSocketClient;

  beforeEach(() => {
    client = new WebSocketClient();
  });

  afterEach(() => {
    client.disconnect();
    vi.clearAllTimers();
  });

  it("should connect and handle messages", async () => {
    vi.useFakeTimers();
    try {
      const handler = vi.fn();
      client.subscribe(handler);
      client.connect();
      await vi.advanceTimersByTimeAsync(10);

      expect(client.isConnected()).toBe(true);

      const message = {
        type: "message",
        data: { message_id: "TEST-001" },
      };

      const ws = getWebSocket(client);
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

  it("should manage subscriptions", () => {
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
});
