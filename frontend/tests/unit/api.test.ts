import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { search } from "../../src/lib/services/api";

vi.mock("../../src/lib/utils/logger", () => ({
  createLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  }),
}));

describe("API Service", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.clearAllMocks();
  });

  it("should fetch search results", async () => {
    const mockData = {
      telegrams: [
        {
          message_id: "TEST-001",
          type: "AFTN",
          time: "2024-01-01T12:00:00Z",
          content: "test",
        },
      ],
      total: 1,
    };

    globalThis.fetch = vi.fn(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockData), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        })
      )
    ) as typeof fetch;

    const result = await search({ query: "test" });
    expect(result.total).toBe(1);
    expect(result.telegrams).toHaveLength(1);
    expect(globalThis.fetch).toHaveBeenCalledTimes(1);
  });
});
