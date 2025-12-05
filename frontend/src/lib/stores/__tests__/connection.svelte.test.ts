import { describe, it, expect } from "vitest";
import { getConnectionStore } from "../connection.svelte";

describe("ConnectionStore", () => {
  it("should initialize with disconnected status", () => {
    const store = getConnectionStore();
    expect(store.httpStatus).toBe("disconnected");
  });

  it("should have no error initially", () => {
    const store = getConnectionStore();
    expect(store.lastError).toBeNull();
  });

  it("should mark HTTP success", () => {
    const store = getConnectionStore();
    store.markHttpSuccess();
    expect(store.httpStatus).toBe("connected");
    expect(store.lastError).toBeNull();
  });

  it("should mark HTTP error", () => {
    const store = getConnectionStore();
    store.markHttpError("Test error");
    // markHttpError sets status to disconnected, then schedules retry which sets it to connecting
    // So status will be either "disconnected" or "connecting" depending on timing
    expect(["disconnected", "connecting"]).toContain(store.httpStatus);
    expect(store.lastError).toBe("Test error");
  });
});

