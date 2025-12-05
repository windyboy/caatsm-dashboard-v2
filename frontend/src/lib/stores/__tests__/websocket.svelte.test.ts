import { describe, it, expect } from "vitest";
import { getWebSocketStore } from "../websocket.svelte";

describe("WebSocketStore", () => {
  it("should initialize with disconnected status", () => {
    const store = getWebSocketStore();
    expect(store.status).toBe("disconnected");
  });

  it("should have null stats initially", () => {
    const store = getWebSocketStore();
    expect(store.stats).toBeNull();
  });

  it("should have empty messages array initially", () => {
    const store = getWebSocketStore();
    expect(store.newMessages).toEqual([]);
  });

  it("should clear messages", () => {
    const store = getWebSocketStore();
    store.clearNewMessages();
    expect(store.newMessages).toEqual([]);
  });
});

