import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";
import { messages } from "../../src/lib/stores/messages";
import { stats } from "../../src/lib/stores/stats";
import type { Telegram } from "../../src/lib/utils/types";

describe("messages store", () => {
  beforeEach(() => {
    messages.clear();
  });

  it("should add a single message", () => {
    const telegram: Telegram = {
      message_id: "TEST-001",
      type: "AFTN",
      time: new Date().toISOString(),
      flight_number: "AA123",
      source: "KJFK",
      destination: "KLAX",
      priority: 1,
      content: "test",
      raw_data: "test",
    };

    messages.add(telegram);
    const storeMessages = get(messages);
    expect(storeMessages).toHaveLength(1);
    expect(storeMessages[0].message_id).toBe("TEST-001");
  });

  it("should add multiple messages", () => {
    const telegram1: Telegram = {
      message_id: "TEST-001",
      type: "AFTN",
      time: new Date().toISOString(),
      flight_number: "AA123",
      source: "KJFK",
      destination: "KLAX",
      priority: 1,
      content: "test",
      raw_data: "test",
    };

    const telegram2: Telegram = {
      message_id: "TEST-002",
      type: "AFTN",
      time: new Date().toISOString(),
      flight_number: "AA456",
      source: "KJFK",
      destination: "KLAX",
      priority: 2,
      content: "test",
      raw_data: "test",
    };

    messages.add(telegram1);
    messages.add(telegram2);

    const storeMessages = get(messages);
    expect(storeMessages).toHaveLength(2);
    expect(storeMessages[0].message_id).toBe("TEST-002");
    expect(storeMessages[1].message_id).toBe("TEST-001");
  });

  it("should enforce MAX_MESSAGES limit", () => {
    const MAX_MESSAGES = 50;

    for (let i = 0; i < MAX_MESSAGES + 10; i++) {
      const telegram: Telegram = {
        message_id: `TEST-${i}`,
        type: "AFTN",
        time: new Date().toISOString(),
        flight_number: "AA123",
        source: "KJFK",
        destination: "KLAX",
        priority: 1,
        content: "test",
        raw_data: "test",
      };
      messages.add(telegram);
    }

    const storeMessages = get(messages);
    expect(storeMessages).toHaveLength(MAX_MESSAGES);
    expect(storeMessages[0].message_id).toBe(`TEST-${MAX_MESSAGES + 9}`);
  });

  it("should add multiple messages at once", () => {
    const telegrams: Telegram[] = Array.from({ length: 5 }, (_, i) => ({
      message_id: `TEST-${i}`,
      type: "AFTN",
      time: new Date().toISOString(),
      flight_number: "AA123",
      source: "KJFK",
      destination: "KLAX",
      priority: 1,
      content: "test",
      raw_data: "test",
    }));

    messages.addMultiple(telegrams);
    const storeMessages = get(messages);
    expect(storeMessages).toHaveLength(5);
  });

  it("should clear messages", () => {
    const telegram: Telegram = {
      message_id: "TEST-001",
      type: "AFTN",
      time: new Date().toISOString(),
      flight_number: "AA123",
      source: "KJFK",
      destination: "KLAX",
      priority: 1,
      content: "test",
      raw_data: "test",
    };

    messages.add(telegram);
    expect(get(messages)).toHaveLength(1);

    messages.clear();
    expect(get(messages)).toHaveLength(0);
  });
});

describe("stats store", () => {
  beforeEach(() => {
    stats.reset();
  });

  it("should set total", () => {
    stats.setTotal(100);
    const storeStats = get(stats);
    expect(storeStats.total).toBe(100);
  });

  it("should set byPriority", () => {
    const byPriority = { 1: 50, 2: 30, 3: 20 };
    stats.setByPriority(byPriority);
    const storeStats = get(stats);
    expect(storeStats.byPriority).toEqual(byPriority);
  });

  it("should set byType", () => {
    const byType = { AFTN: 50, ACARS: 30 };
    stats.setByType(byType);
    const storeStats = get(stats);
    expect(storeStats.byType).toEqual(byType);
  });

  it("should reset to initial state", () => {
    stats.setTotal(100);
    stats.setByPriority({ 1: 50 });
    stats.setByType({ AFTN: 50 });

    stats.reset();
    const storeStats = get(stats);
    expect(storeStats.total).toBe(0);
    expect(storeStats.byPriority).toEqual({});
    expect(storeStats.byType).toEqual({});
  });

  it("should update multiple fields independently", () => {
    stats.setTotal(100);
    stats.setByPriority({ 1: 50 });
    stats.setByType({ AFTN: 50 });

    const storeStats = get(stats);
    expect(storeStats.total).toBe(100);
    expect(storeStats.byPriority).toEqual({ 1: 50 });
    expect(storeStats.byType).toEqual({ AFTN: 50 });
  });
});
