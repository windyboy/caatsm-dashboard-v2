import { describe, it, expect, vi, beforeEach } from "vitest";
import { fetchStats, fetchHealth, runSearch } from "../api";

// Mock fetch globally
global.fetch = vi.fn();

describe("API Client", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should build URLs correctly", async () => {
    const mockResponse = { telegrams: [], total: 0 };
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    });

    await runSearch("test");
    expect(global.fetch).toHaveBeenCalled();
  });

  it("should handle fetch errors", async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new TypeError("fetch failed"));

    await expect(runSearch("test")).rejects.toThrow();
  });

  it("should handle non-ok responses", async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      json: async () => ({}),
    });

    await expect(runSearch("test")).rejects.toThrow();
  });
});
