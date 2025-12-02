import { beforeEach, describe, expect, it } from "vitest";
import { get } from "svelte/store";
import { messages } from "$lib/stores/messages";
import { stats } from "$lib/stores/stats";
import type { Telegram } from "$lib/utils/types";

describe("messages store", () => {
  beforeEach(() => {
    messages.clear();
  });

  it("should add and clear messages", () => {
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
    expect(get(messages)[0].message_id).toBe("TEST-001");

    messages.clear();
    expect(get(messages)).toHaveLength(0);
  });
});

describe("stats store", () => {
  beforeEach(() => {
    stats.reset();
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
});
