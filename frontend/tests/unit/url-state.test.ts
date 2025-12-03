import { describe, it, expect } from "vitest";
import type { SearchParams } from "$lib/services/api";

describe("URL state serialization/deserialization", () => {
  it("should serialize SearchParams to URLSearchParams correctly", () => {
    const params: SearchParams = {
      query: "test query",
      type: "AFTN",
      priority: 1,
      start_time: "2024-01-01T00:00:00Z",
      end_time: "2024-01-02T00:00:00Z",
      limit: 50,
      offset: 0,
      sort_by: "time",
      order: "desc",
    };

    const searchParams = new URLSearchParams();
    if (params.query) searchParams.set("query", params.query);
    if (params.type) searchParams.set("type", params.type);
    if (params.priority !== undefined) searchParams.set("priority", params.priority.toString());
    if (params.start_time) searchParams.set("start_time", params.start_time);
    if (params.end_time) searchParams.set("end_time", params.end_time);
    if (params.limit !== undefined) searchParams.set("limit", params.limit.toString());
    if (params.offset !== undefined) searchParams.set("offset", params.offset.toString());
    if (params.sort_by) searchParams.set("sort_by", params.sort_by);
    if (params.order) searchParams.set("order", params.order);

    expect(searchParams.get("query")).toBe("test query");
    expect(searchParams.get("type")).toBe("AFTN");
    expect(searchParams.get("priority")).toBe("1");
    expect(searchParams.get("start_time")).toBe("2024-01-01T00:00:00Z");
    expect(searchParams.get("end_time")).toBe("2024-01-02T00:00:00Z");
    expect(searchParams.get("limit")).toBe("50");
    expect(searchParams.get("offset")).toBe("0");
    expect(searchParams.get("sort_by")).toBe("time");
    expect(searchParams.get("order")).toBe("desc");
  });

  it("should deserialize URLSearchParams to SearchParams correctly", () => {
    const searchParams = new URLSearchParams();
    searchParams.set("query", "test");
    searchParams.set("type", "SITA");
    searchParams.set("priority", "2");
    searchParams.set("start_time", "2024-01-01T00:00:00Z");
    searchParams.set("end_time", "2024-01-02T00:00:00Z");

    const params: SearchParams = {
      query: searchParams.get("query") || undefined,
      type: searchParams.get("type") || undefined,
      priority: searchParams.get("priority") ? parseInt(searchParams.get("priority")!) : undefined,
      start_time: searchParams.get("start_time") || undefined,
      end_time: searchParams.get("end_time") || undefined,
    };

    expect(params.query).toBe("test");
    expect(params.type).toBe("SITA");
    expect(params.priority).toBe(2);
    expect(params.start_time).toBe("2024-01-01T00:00:00Z");
    expect(params.end_time).toBe("2024-01-02T00:00:00Z");
  });

  it("should handle missing optional parameters", () => {
    const searchParams = new URLSearchParams();
    searchParams.set("query", "test");

    const params: SearchParams = {
      query: searchParams.get("query") || undefined,
      type: searchParams.get("type") || undefined,
      priority: searchParams.get("priority") ? parseInt(searchParams.get("priority")!) : undefined,
    };

    expect(params.query).toBe("test");
    expect(params.type).toBeUndefined();
    expect(params.priority).toBeUndefined();
  });
});

